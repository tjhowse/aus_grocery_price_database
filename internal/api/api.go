// Package api serves product data in the format used by Heisse Preise:
// https://github.com/badlogic/heissepreise
package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tjhowse/aus_grocery_price_database/internal/shared"
	"golang.org/x/time/rate"
)

const DEFAULT_PAGE_SIZE = 100
const MAX_PAGE_SIZE = 1000
const PRICE_HISTORY_TIMEOUT = 10 * time.Second

// Each history request is a query over a product's whole lifetime in InfluxDB, which is
// short on RAM. Requests beyond this many in flight get a 503 rather than queueing.
const MAX_CONCURRENT_HISTORY_QUERIES = 4

// Per client IP. A burst of 20 lets a page load several products at once.
const RATE_LIMIT_PER_SECOND = 2
const RATE_LIMIT_BURST = 20
const RATE_LIMIT_IDLE_EXPIRY = 10 * time.Minute

// Prices are re-scraped every few hours, so an hour of caching costs little freshness.
const CACHE_CONTROL = "public, max-age=3600"

// validID matches the product IDs used by the stores, and rejects anything that
// would need escaping in a database query.
var validID = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

// ProductSource provides product information from a store's local database.
type ProductSource interface {
	GetSharedProductsPage(afterID string, count int) ([]shared.ProductInfo, error)
	GetSharedProduct(id string) (shared.ProductInfo, error)
}

// PriceHistorian provides the price history of a product by its shared ID.
type PriceHistorian interface {
	GetPriceHistory(ctx context.Context, id string) ([]shared.PricePoint, error)
}

// Store is a grocery store exposed by the API.
type Store struct {
	Name     string // Used in URLs and the "store" field, E.G. "woolworths"
	IDPrefix string // Prepended to the store's own product ID to make the shared ID
	Source   ProductSource
}

// PricePoint is a price change in the Heisse Preise format.
type PricePoint struct {
	Date  string  `json:"date"`
	Price float64 `json:"price"`
}

// Product is a product in the Heisse Preise format.
type Product struct {
	Store        string       `json:"store"`
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Price        float64      `json:"price"`
	PriceHistory []PricePoint `json:"priceHistory"`
	IsWeighted   bool         `json:"isWeighted"`
	Unit         string       `json:"unit,omitempty"`
	Quantity     int          `json:"quantity,omitempty"`
	Bio          bool         `json:"bio"`
	Category     string       `json:"category,omitempty"`
}

type Server struct {
	stores       map[string]Store
	historian    PriceHistorian
	historySlots chan struct{}
	limiter      *ipRateLimiter
}

// NewServer creates an API server for the given stores. historian may be nil,
// in which case every priceHistory is empty.
func NewServer(stores []Store, historian PriceHistorian) *Server {
	s := Server{
		stores:       make(map[string]Store),
		historian:    historian,
		historySlots: make(chan struct{}, MAX_CONCURRENT_HISTORY_QUERIES),
		limiter:      newIPRateLimiter(RATE_LIMIT_PER_SECOND, RATE_LIMIT_BURST),
	}
	for _, store := range stores {
		s.stores[store.Name] = store
	}
	return &s
}

// Handler returns the API's routes, rate limited and logged.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/products/{store}", s.handleProductList)
	mux.HandleFunc("GET /api/v1/products/{store}/{id}", s.handleProduct)
	return logRequests(s.rateLimit(mux))
}

// rateLimit rejects requests from clients over their rate limit.
func (s *Server) rateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.limiter.allow(clientIP(r)) {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// statusRecorder captures the response status for logging.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// logRequests logs one line per request.
func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		slog.Info("API request",
			"method", r.Method,
			"path", r.URL.Path,
			"query", r.URL.RawQuery,
			"status", rec.status,
			"duration", time.Since(start),
			"ip", clientIP(r),
			"userAgent", r.UserAgent())
	})
}

// clientIP returns the requesting client's IP. Fly's proxy sets Fly-Client-IP and
// overwrites any value the client sends, so it can be trusted behind Fly.
func clientIP(r *http.Request) string {
	if ip := r.Header.Get("Fly-Client-IP"); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// ipRateLimiter keeps a token bucket per client IP, forgetting idle clients.
type ipRateLimiter struct {
	mu        sync.Mutex
	visitors  map[string]*visitor
	perSec    rate.Limit
	burst     int
	lastSweep time.Time
}

func newIPRateLimiter(perSec float64, burst int) *ipRateLimiter {
	return &ipRateLimiter{visitors: make(map[string]*visitor), perSec: rate.Limit(perSec), burst: burst}
}

func (l *ipRateLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	// Forget idle clients at most once a minute, so the map can't grow without bound.
	if now.Sub(l.lastSweep) > time.Minute {
		for k, old := range l.visitors {
			if now.Sub(old.lastSeen) > RATE_LIMIT_IDLE_EXPIRY {
				delete(l.visitors, k)
			}
		}
		l.lastSweep = now
	}
	v, ok := l.visitors[ip]
	if !ok {
		v = &visitor{limiter: rate.NewLimiter(l.perSec, l.burst)}
		l.visitors[ip] = v
	}
	v.lastSeen = now
	return v.limiter.Allow()
}

// ListenAndServe serves the API on addr until it fails.
func (s *Server) ListenAndServe(addr string) error {
	slog.Info("Serving API", "address", addr)
	server := http.Server{
		Addr:              addr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
	}
	return server.ListenAndServe()
}

// handleProductList returns a page of a store's products, ordered by ID, without price history.
// Request the next page by passing the last product's ID as "after".
func (s *Server) handleProductList(w http.ResponseWriter, r *http.Request) {
	store, ok := s.stores[r.PathValue("store")]
	if !ok {
		http.Error(w, "unknown store", http.StatusNotFound)
		return
	}
	after := r.URL.Query().Get("after")
	if after != "" && !validID.MatchString(after) {
		http.Error(w, "invalid after", http.StatusBadRequest)
		return
	}
	limit := DEFAULT_PAGE_SIZE
	if l := r.URL.Query().Get("limit"); l != "" {
		var err error
		limit, err = strconv.Atoi(l)
		if err != nil || limit < 1 || limit > MAX_PAGE_SIZE {
			http.Error(w, "limit must be between 1 and "+strconv.Itoa(MAX_PAGE_SIZE), http.StatusBadRequest)
			return
		}
	}
	infos, err := store.Source.GetSharedProductsPage(after, limit)
	if err != nil {
		slog.Error("Error getting product page", "store", store.Name, "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	products := make([]Product, 0, len(infos))
	for _, info := range infos {
		products = append(products, toProduct(store, info, nil))
	}
	writeJSON(w, products)
}

// handleProduct returns a single product with its price history.
func (s *Server) handleProduct(w http.ResponseWriter, r *http.Request) {
	store, ok := s.stores[r.PathValue("store")]
	if !ok {
		http.Error(w, "unknown store", http.StatusNotFound)
		return
	}
	id := r.PathValue("id")
	if !validID.MatchString(id) {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	info, err := store.Source.GetSharedProduct(id)
	if errors.Is(err, shared.ErrProductMissing) {
		http.Error(w, "product not found", http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("Error getting product", "store", store.Name, "id", id, "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	var history []shared.PricePoint
	if s.historian != nil {
		select {
		case s.historySlots <- struct{}{}:
			defer func() { <-s.historySlots }()
		default:
			w.Header().Set("Retry-After", "5")
			http.Error(w, "too many price history requests, try again shortly", http.StatusServiceUnavailable)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), PRICE_HISTORY_TIMEOUT)
		defer cancel()
		history, err = s.historian.GetPriceHistory(ctx, info.ID)
		if err != nil {
			slog.Error("Error getting price history", "id", info.ID, "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}
	writeJSON(w, toProduct(store, info, history))
}

// toProduct converts a product to the Heisse Preise format. history is oldest first.
func toProduct(store Store, info shared.ProductInfo, history []shared.PricePoint) Product {
	p := Product{
		Store:        store.Name,
		ID:           strings.TrimPrefix(info.ID, store.IDPrefix),
		Name:         info.Name,
		Description:  info.Description,
		Price:        centsToDollars(info.PriceCents),
		PriceHistory: priceChanges(history),
		Category:     info.Department,
	}
	if info.WeightGrams > 0 {
		p.Unit = "g"
		p.Quantity = info.WeightGrams
	}
	return p
}

// priceChanges reduces an oldest-first price history to the points where the price
// changed, newest first, as Heisse Preise does. Zero prices are missing data and are skipped.
func priceChanges(history []shared.PricePoint) []PricePoint {
	changes := []PricePoint{}
	lastCents := 0
	for _, point := range history {
		if point.PriceCents == 0 || point.PriceCents == lastCents {
			continue
		}
		lastCents = point.PriceCents
		changes = append(changes, PricePoint{
			Date:  point.Time.UTC().Format(time.DateOnly),
			Price: centsToDollars(point.PriceCents),
		})
	}
	for i, j := 0, len(changes)-1; i < j; i, j = i+1, j-1 {
		changes[i], changes[j] = changes[j], changes[i]
	}
	return changes
}

func centsToDollars(cents int) float64 {
	return float64(cents) / 100
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", CACHE_CONTROL)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		slog.Error("Error writing JSON response", "error", err)
	}
}
