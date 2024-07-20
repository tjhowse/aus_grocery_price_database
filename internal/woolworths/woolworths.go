package woolworths

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/time/rate"
)

const WOOLWORTHS_PRODUCT_URL_FORMAT = "%s/api/v3/ui/schemaorg/product/%d"
const DB_SCHEMA_VERSION = 1
const PRODUCT_INFO_WORKER_COUNT = 5

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
		Ratelimiter: rate.NewLimiter(rate.Every(1*time.Second), 1),
	}
	w.productMaxAge = productMaxAge
	err = w.InitDB(dbPath)
	if err != nil {
		return err
	}
	return nil
}

// Runs up all the workers and mediates data flowing between them.
// Currently all sqlite writes happen via this function. This may move
// off to a separate goroutine in the future.
func (w *Woolworths) RunScheduler(cancel chan struct{}) {

	productInfoChannel := make(chan WoolworthsProductInfo)
	productsThatNeedAnUpdateChannel := make(chan ProductID)
	newDepartmentIDsChannel := make(chan DepartmentID)
	for i := 0; i < PRODUCT_INFO_WORKER_COUNT; i++ {
		go w.ProductInfoFetchingWorker(productsThatNeedAnUpdateChannel, productInfoChannel)
	}
	go w.ProductUpdateQueueWorker(productsThatNeedAnUpdateChannel, w.productMaxAge)
	go w.NewProductWorker(productInfoChannel)
	go w.NewDepartmentIDWorker(newDepartmentIDsChannel)

	for {
		slog.Debug("Heartbeat")
		select {
		case productInfoUpdate := <-productInfoChannel:
			slog.Debug("Read from productInfoChannel", "name", productInfoUpdate.Info.Name)
			// Update the product info in the DB
			err := w.SaveProductInfo(productInfoUpdate)
			if err != nil {
				slog.Error(fmt.Sprintf("Error saving product info: %v", err))
			}
		case newDepartmentID := <-newDepartmentIDsChannel:
			slog.Debug(fmt.Sprintf("New department ID: %s", newDepartmentID))
			// Update the departmentIDs table with the new department ID
			err := w.SaveDepartment(newDepartmentID)
			if err != nil {
				slog.Error(fmt.Sprintf("Error saving department ID: %v", err))
			}
		case <-cancel:
			slog.Info("Exiting scheduler")
			return
		}
	}

}
