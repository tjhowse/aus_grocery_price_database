package coles

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/tjhowse/aus_grocery_price_database/internal/shared"
	"golang.org/x/time/rate"
)

const DEFAULT_LISTING_PAGE_CHECK_INTERVAL = 1 * time.Minute

// Coles satisfies the ProductInfoGetter interface.
type Coles struct {
	baseURL                   string
	client                    *shared.RLHTTPClient
	cookieJar                 *cookiejar.Jar // TODO This might not be threadsafe.
	db                        *sql.DB
	colesAPIVersion           string
	productMaxAge             time.Duration
	listingPageUpdateInterval time.Duration
	filteredDepartmentIDsSet  map[string]bool
	filterDepartments         bool
	logger                    *slog.Logger
}

// Init initialises the Coles struct.
func (c *Coles) Init(baseURL string, dbPath string, productMaxAge time.Duration) error {
	var err error

	c.logger = slog.With("store", "Coles")

	// This might change on occasion. We should allow for that.
	//'https://www.coles.com.au/_next/data/20240809.03_v4.7.3/en/browse.json'
	c.colesAPIVersion = DEFAULT_API_VERSION
	c.baseURL = baseURL

	c.cookieJar, err = cookiejar.New(nil)
	if err != nil {
		return fmt.Errorf("error creating cookie jar: %v", err)
	}
	c.baseURL = baseURL
	c.client = &shared.RLHTTPClient{
		Client: &http.Client{
			Jar:     c.cookieJar,
			Timeout: 30 * time.Second,
		},
		Ratelimiter: rate.NewLimiter(rate.Every(1000*time.Millisecond), 1),
	}
	c.productMaxAge = productMaxAge
	err = c.initDB(dbPath)
	if err != nil {
		return err
	}
	c.listingPageUpdateInterval = DEFAULT_LISTING_PAGE_CHECK_INTERVAL
	c.filteredDepartmentIDsSet = map[string]bool{
		"fruit-vegetables":  true,
		"dairy-eggs-fridge": true,
		"bakery":            true,
		"deli":              true,
		"pantry":            true,
		"meat-seafood":      true,
		"frozen":            true,
		"drinks":            true,
		"household":         true,
		// "health-beauty":     true,
		// "baby":              true,
		// "pet":               true,
		// "liquor":            true,
		// "tobacco":           true,
	}
	c.filterDepartments = true

	if err := c.updateAPIVersion(); err != nil {
		slog.Error("error updating API version", "error", err)
	}

	return nil
}

// Runs up all the workers and mediates data flowing between them.
// Currently all sqlite writes happen via this function. This may move
// off to a separate goroutine in the future.
func (c *Coles) Run(cancel chan struct{}) {
	departmentPageChannel := make(chan departmentPage)

	go c.productListPageWorker(departmentPageChannel)
	go c.newDepartmentInfoWorker()
	go c.departmentPageUpdateQueueWorker(departmentPageChannel, c.productMaxAge)

	for range cancel {
		return
	}
}

// GetSharedProductsUpdatedAfter provides a list of product IDs that have been updated since the given time
func (c *Coles) GetSharedProductsUpdatedAfter(t time.Time, count int) ([]shared.ProductInfo, error) {
	return c.querySharedProducts("products.updated > ? AND name != '' LIMIT ?", t, count)
}

// GetSharedProductsPage provides up to count products with an ID greater than afterID, ordered by ID.
// afterID is the store's own product ID, without the shared prefix. Pass "" for the first page.
func (c *Coles) GetSharedProductsPage(afterID string, count int) ([]shared.ProductInfo, error) {
	return c.querySharedProducts("productID > ? AND name != '' ORDER BY productID LIMIT ?", afterID, count)
}

// GetSharedProduct provides a single product by the store's own product ID, without the shared prefix.
func (c *Coles) GetSharedProduct(id string) (shared.ProductInfo, error) {
	products, err := c.querySharedProducts("productID = ? AND name != ''", id)
	if err != nil {
		return shared.ProductInfo{}, err
	}
	if len(products) == 0 {
		return shared.ProductInfo{}, shared.ErrProductMissing
	}
	return products[0], nil
}

// querySharedProducts reads products matching the where clause into the shared format.
func (c *Coles) querySharedProducts(where string, args ...any) ([]shared.ProductInfo, error) {
	var productIDs []shared.ProductInfo
	var deptDescription sql.NullString
	rows, err := c.db.Query(`
		SELECT
			productID,
			products.name,
			products.description,
			departments.description,
			priceCents,
			previousPriceCents,
			weightGrams,
			products.updated
		FROM
			products
			LEFT JOIN departments ON products.departmentID = departments.departmentID
		WHERE `+where, args...)
	if err != nil {
		return productIDs, fmt.Errorf("failed to query productIDs: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var product shared.ProductInfo
		err = rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&deptDescription,
			&product.PriceCents,
			&product.PreviousPriceCents,
			&product.WeightGrams,
			&product.Timestamp)
		if err != nil {
			return productIDs, fmt.Errorf("failed to scan productID: %w", err)
		}
		if deptDescription.Valid {
			product.Department = deptDescription.String
		}
		product.ID = COLES_ID_PREFIX + product.ID
		product.Store = "Coles"
		productIDs = append(productIDs, product)
	}
	return productIDs, nil
}

// GetTotalProductCount returns the total number of products in the database.
func (c *Coles) GetTotalProductCount() (int, error) {
	var count int
	err := c.db.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to query product count: %w", err)
	}
	return count, nil
}
