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
	baseURL       string
	client        *RLHTTPClient
	cookieJar     *cookiejar.Jar // TODO This might not be threadsafe.
	db            *sql.DB
	productMaxAge time.Duration
}

// Returns a list of product IDs that have been updated since the given time
func (w *Woolworths) GetSharedProductsUpdatedAfter(t time.Time, count int) ([]shared.ProductInfo, error) {
	var productIDs []shared.ProductInfo
	rows, err := w.db.Query("SELECT productID, name, description, priceCents, weightGrams, updated FROM products WHERE updated > ? LIMIT ?", t, count)
	if err != nil {
		return productIDs, fmt.Errorf("failed to query productIDs: %w", err)
	}
	for rows.Next() {
		var product shared.ProductInfo
		err = rows.Scan(&product.ID, &product.Name, &product.Description, &product.PriceCents, &product.WeightGrams, &product.Timestamp)
		if err != nil {
			return productIDs, fmt.Errorf("failed to scan productID: %w", err)
		}
		product.ID = WOOLWORTHS_ID_PREFIX + product.ID
		productIDs = append(productIDs, product)
	}
	return productIDs, nil
}

func (w *Woolworths) Init(baseURL string, dbPath string, productMaxAge time.Duration) error {
	var err error
	w.cookieJar, err = cookiejar.New(nil)
	if err != nil {
		return fmt.Errorf("error creating cookie jar: %v", err)
	}
	w.baseURL = baseURL
	w.client = &RLHTTPClient{
		client: &http.Client{
			Jar: w.cookieJar,
		},
		Ratelimiter: rate.NewLimiter(rate.Every(50*time.Millisecond), 1),
	}
	w.productMaxAge = productMaxAge
	err = w.initDB(dbPath)
	if err != nil {
		return err
	}
	return nil
}
