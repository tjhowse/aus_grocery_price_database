package coles

import (
	"database/sql"
	"fmt"
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
}

// Init initialises the Coles struct.
func (c *Coles) Init(baseURL string, dbPath string, productMaxAge time.Duration) error {
	var err error
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
	return nil
}

// Run starts the Coles product information getter.
func (c *Coles) Run(stop <-chan struct{}) {
}

// GetSharedProductsUpdatedAfter provides a list of product IDs that have been updated since the given time
func (c *Coles) GetSharedProductsUpdatedAfter(t int, count int) ([]shared.ProductInfo, error) {
	return nil, nil
}

// GetTotalProductCount returns the total number of products in the database.
func (c *Coles) GetTotalProductCount() (int, error) {
	return 0, nil
}
