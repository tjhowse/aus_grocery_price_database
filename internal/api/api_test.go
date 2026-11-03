package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/tjhowse/aus_grocery_price_database/internal/shared"
)

type mockSource struct {
	products  []shared.ProductInfo
	lastAfter string
	lastCount int
}

func (m *mockSource) GetSharedProductsPage(afterID string, count int) ([]shared.ProductInfo, error) {
	m.lastAfter = afterID
	m.lastCount = count
	return m.products, nil
}

func (m *mockSource) GetSharedProduct(id string) (shared.ProductInfo, error) {
	for _, p := range m.products {
		if p.ID == "test_"+id {
			return p, nil
		}
	}
	return shared.ProductInfo{}, shared.ErrProductMissing
}

type mockHistorian struct {
	history []shared.PricePoint
	lastID  string
}

func (m *mockHistorian) GetPriceHistory(ctx context.Context, id string) ([]shared.PricePoint, error) {
	m.lastID = id
	return m.history, nil
}

func day(d int) time.Time {
	return time.Date(2025, 1, d, 6, 0, 0, 0, time.UTC)
}

func getTestServer() (*Server, *mockSource, *mockHistorian) {
	source := &mockSource{products: []shared.ProductInfo{
		{ID: "test_123", Name: "Navel Oranges", Description: "Each", Department: "Fruit & Veg", PriceCents: 649, WeightGrams: 200},
		{ID: "test_456", Name: "Cloth Nappy", PriceCents: 1099},
	}}
	historian := &mockHistorian{history: []shared.PricePoint{
		{Time: day(1), PriceCents: 1099},
		{Time: day(2), PriceCents: 1099},
		{Time: day(3), PriceCents: 0},
		{Time: day(4), PriceCents: 899},
		{Time: day(5), PriceCents: 649},
	}}
	return NewServer([]Store{{Name: "test", IDPrefix: "test_", Source: source}}, historian), source, historian
}

func get(t *testing.T, s *Server, url string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, url, nil))
	return rec
}

func TestPriceChanges(t *testing.T) {
	_, _, historian := getTestServer()
	want := []PricePoint{
		{Date: "2025-01-05", Price: 6.49},
		{Date: "2025-01-04", Price: 8.99},
		{Date: "2025-01-01", Price: 10.99},
	}
	if got := priceChanges(historian.history); !reflect.DeepEqual(want, got) {
		t.Errorf("Expected %v, got %v", want, got)
	}
	if got := priceChanges(nil); got == nil || len(got) != 0 {
		t.Errorf("Expected an empty list, got %v", got)
	}
}

func TestProduct(t *testing.T) {
	s, _, historian := getTestServer()
	rec := get(t, s, "/api/v1/products/test/123")
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got Product
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	want := Product{
		Store:       "test",
		ID:          "123",
		Name:        "Navel Oranges",
		Description: "Each",
		Price:       6.49,
		PriceHistory: []PricePoint{
			{Date: "2025-01-05", Price: 6.49},
			{Date: "2025-01-04", Price: 8.99},
			{Date: "2025-01-01", Price: 10.99},
		},
		Unit:     "g",
		Quantity: 200,
		Category: "Fruit & Veg",
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("Expected %+v, got %+v", want, got)
	}
	if want, got := "test_123", historian.lastID; want != got {
		t.Errorf("Expected history lookup by %s, got %s", want, got)
	}
}

func TestProductList(t *testing.T) {
	s, source, _ := getTestServer()
	rec := get(t, s, "/api/v1/products/test?after=100&limit=2")
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got []Product
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("Expected 2 products, got %d", len(got))
	}
	if got[1].ID != "456" || got[1].Unit != "" || got[1].PriceHistory == nil {
		t.Errorf("Unexpected product %+v", got[1])
	}
	if source.lastAfter != "100" || source.lastCount != 2 {
		t.Errorf("Expected after=100 count=2, got after=%s count=%d", source.lastAfter, source.lastCount)
	}
	// An empty priceHistory must be [] rather than null.
	var raw []map[string]any
	json.Unmarshal(rec.Body.Bytes(), &raw)
	if _, ok := raw[0]["priceHistory"].([]any); !ok {
		t.Errorf("Expected priceHistory to be a list, got %v", raw[0]["priceHistory"])
	}
}

func TestNoHistorian(t *testing.T) {
	s, source, _ := getTestServer()
	s = NewServer([]Store{{Name: "test", IDPrefix: "test_", Source: source}}, nil)
	rec := get(t, s, "/api/v1/products/test/123")
	var got Product
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.PriceHistory == nil || len(got.PriceHistory) != 0 {
		t.Errorf("Expected an empty priceHistory, got %v", got.PriceHistory)
	}
}

func TestErrors(t *testing.T) {
	s, _, _ := getTestServer()
	for url, want := range map[string]int{
		"/api/v1/products/aldi":              http.StatusNotFound,
		"/api/v1/products/aldi/123":          http.StatusNotFound,
		"/api/v1/products/test/999":          http.StatusNotFound,
		"/api/v1/products/test/12%22%29":     http.StatusBadRequest,
		"/api/v1/products/test?after=a%22b":  http.StatusBadRequest,
		"/api/v1/products/test?limit=0":      http.StatusBadRequest,
		"/api/v1/products/test?limit=100000": http.StatusBadRequest,
		"/api/v1/products/test?limit=abc":    http.StatusBadRequest,
	} {
		if got := get(t, s, url).Code; got != want {
			t.Errorf("%s: expected %d, got %d", url, want, got)
		}
	}
}

func TestRateLimit(t *testing.T) {
	s, _, _ := getTestServer()
	for i := 0; i < RATE_LIMIT_BURST; i++ {
		if got := get(t, s, "/api/v1/products/test").Code; got != http.StatusOK {
			t.Fatalf("Request %d: expected 200, got %d", i, got)
		}
	}
	if want, got := http.StatusTooManyRequests, get(t, s, "/api/v1/products/test").Code; want != got {
		t.Errorf("Expected %d, got %d", want, got)
	}
	// Another client is unaffected.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/test", nil)
	req.Header.Set("Fly-Client-IP", "203.0.113.9")
	s.Handler().ServeHTTP(rec, req)
	if want, got := http.StatusOK, rec.Code; want != got {
		t.Errorf("Expected %d, got %d", want, got)
	}
}

func TestHistoryConcurrencyLimit(t *testing.T) {
	s, _, _ := getTestServer()
	for i := 0; i < MAX_CONCURRENT_HISTORY_QUERIES; i++ {
		s.historySlots <- struct{}{}
	}
	if want, got := http.StatusServiceUnavailable, get(t, s, "/api/v1/products/test/123").Code; want != got {
		t.Errorf("Expected %d, got %d", want, got)
	}
	<-s.historySlots
	if want, got := http.StatusOK, get(t, s, "/api/v1/products/test/123").Code; want != got {
		t.Errorf("Expected %d, got %d", want, got)
	}
}
