package woolworths

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/tjhowse/aus_grocery_price_database/internal/shared"
	"golang.org/x/time/rate"
)

const WOOLWORTHS_PRODUCT_URL_FORMAT = "%s/api/v3/ui/schemaorg/product/%s"
const PRODUCT_INFO_WORKER_COUNT = 10

const PRODUCT_INFO_MAX_AGE = 6 * time.Hour

var ErrProductMissing = errors.New("no product found")

type Woolworths struct {
	baseURL                  string
	client                   *RLHTTPClient
	cookieJar                *cookiejar.Jar // TODO This might not be threadsafe.
	db                       *sql.DB
	productMaxAge            time.Duration
	filterDepartments        bool // These are used to limit the departments and products for gradual testing.
	filterProducts           bool
	filteredDepartmentIDsSet map[departmentID]bool
}

// Returns a list of product IDs that have been updated since the given time
func (w *Woolworths) GetSharedProductsUpdatedAfter(t time.Time, count int) ([]shared.ProductInfo, error) {
	var productIDs []shared.ProductInfo
	rows, err := w.db.Query("SELECT productID, name, description, departmentDescription, priceCents, weightGrams, updated FROM products WHERE updated > ? AND name != '' LIMIT ?", t, count)
	if err != nil {
		return productIDs, fmt.Errorf("failed to query productIDs: %w", err)
	}
	for rows.Next() {
		var product shared.ProductInfo
		err = rows.Scan(&product.ID, &product.Name, &product.Description, &product.Department, &product.PriceCents, &product.WeightGrams, &product.Timestamp)
		if err != nil {
			return productIDs, fmt.Errorf("failed to scan productID: %w", err)
		}
		product.ID = WOOLWORTHS_ID_PREFIX + product.ID
		product.Store = "Woolworths"
		productIDs = append(productIDs, product)
	}
	return productIDs, nil
}

func (w *Woolworths) Init(baseURL string, dbPath string, productMaxAge time.Duration) error {
	var err error

	// These are overridden by tests for now.
	w.filterDepartments = true
	w.filterProducts = false

	w.cookieJar, err = cookiejar.New(nil)
	if err != nil {
		return fmt.Errorf("error creating cookie jar: %v", err)
	}
	w.baseURL = baseURL
	w.client = &RLHTTPClient{
		client: &http.Client{
			Jar:     w.cookieJar,
			Timeout: 30 * time.Second,
		},
		Ratelimiter: rate.NewLimiter(rate.Every(50*time.Millisecond), 1),
	}
	w.productMaxAge = productMaxAge
	err = w.initDB(dbPath)
	if err != nil {
		return err
	}
	w.filteredDepartmentIDsSet = map[departmentID]bool{"1-E5BEE36E": true, "1_DEB537E": true, "1_D5A2236": true}
	return nil
}
