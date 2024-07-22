package woolworths

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"time"

	_ "github.com/mattn/go-sqlite3"
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
