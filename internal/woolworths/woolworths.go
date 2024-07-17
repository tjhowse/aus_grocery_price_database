package woolworths

import (
	"database/sql"
	"encoding/json"
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
const WORKER_COUNT = 5

// const PRODUCT_INFO_MAX_AGE_SECONDS = 24 * time.Hour // 24 hours
const PRODUCT_INFO_MAX_AGE_SECONDS = 10 * time.Second

type Woolworths struct {
	baseURL   string
	client    *RLHTTPClient
	cookieJar *cookiejar.Jar // TODO This might not be threadsafe.
	db        *sql.DB
}

func (w *Woolworths) ProductWorker(input chan ProductID, output chan WoolworthsProductInfo) {
	slog.Debug("Running a Woolworths.Worker")
	for id := range input {
		slog.Debug(fmt.Sprintf("Getting product info for ID: %d", id))
		info, err := w.GetProductInfo(id)
		if err != nil {
			slog.Error(fmt.Sprintf("Error getting product info: %v", err))
		}
		slog.Debug("ProductWorker Blocking on sending out")
		output <- info
		slog.Debug("ProductWorker unblocked on sending out")
	}
}

// Initialises the DB with the schema. Note you must bump the DB_SCHEMA_VERSION
// constant if you change the schema.
func (w *Woolworths) InitBlankDB() {
	w.db.Exec("CREATE TABLE IF NOT EXISTS schema (version INTEGER PRIMARY KEY)")
	w.db.Exec("INSERT INTO schema (version) VALUES (?)", DB_SCHEMA_VERSION)
	w.db.Exec("CREATE TABLE IF NOT EXISTS departmentIDs (departmentID TEXT UNIQUE, retrieved DATETIME)")
	w.db.Exec("CREATE TABLE IF NOT EXISTS productIDs (productID INTEGER UNIQUE, retrieved DATETIME)")
	w.db.Exec("CREATE TABLE IF NOT EXISTS products (productID INTEGER UNIQUE, productData TEXT, retrieved DATETIME)")
}

func (w *Woolworths) InitDB(dbPath string) {
	var err error
	w.db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		slog.Error(fmt.Sprintf("Error opening database: %v", err))
	}
	var version int
	err = w.db.QueryRow("SELECT version FROM schema").Scan(&version)
	if err != nil || version != DB_SCHEMA_VERSION {
		slog.Warn("DB schema error. Creating blank DB.", "version", version)
		w.InitBlankDB()
	}

	// w.db.Exec("CREATE TABLE IF NOT EXISTS product_info (id INTEGER PRIMARY KEY, data TEXT)")
	/*
		Structure the DB as follows:
		Table for product ID
		Table for product info

		Each row in each table has a retrieved datetime field.
		In code, define a maximum age for each type of data.
		The runner will check the age of the data in the DB and refresh it if it's too old.
		If a new product is found, it will be added to the DB and have its datatime set to 0.
		The scheduler queries the DB sorted by datetime and launches a worker to refresh the data
		as required. The workers don't write to the DB, they just return the data to the scheduler.
	*/
}

// Saves product info to the database
func (w *Woolworths) SaveProductInfo(productInfo WoolworthsProductInfo, updateTime time.Time) error {
	var err error
	var result sql.Result

	var productInfoBytes []byte

	slog.Debug("Saving product", "info", productInfo.Info.Description)

	productInfoBytes, err = json.Marshal(productInfo.Info)
	if err != nil {
		return fmt.Errorf("failed to marshal product info: %w", err)
	}
	productInfoString := string(productInfoBytes)

	result, err = w.db.Exec(`
		INSERT INTO products (productID, productData, retrieved)
		VALUES (?, ?, ?)
		ON CONFLICT(productID) DO UPDATE SET productID = ?, productData = ?, retrieved = ?
		`, productInfo.ID, productInfoString, updateTime, productInfo.ID, productInfoString, updateTime)

	if err != nil {
		return fmt.Errorf("failed to update product info: %w", err)
	}
	if rowsAffected, err := result.RowsAffected(); err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	} else if rowsAffected == 0 {
		slog.Warn("Product info not updated.")
	}

	return nil
}

// Loads cached product info from the database
func (w *Woolworths) LoadProductInfo(productID ProductID) (ProductInfo, error) {
	var buffer string
	var result ProductInfo
	err := w.db.QueryRow("SELECT productData FROM products WHERE productID = ? LIMIT 1", productID).Scan(&buffer)
	if err != nil {
		return result, fmt.Errorf("failed to query existing productData: %w", err)
	}
	err = json.Unmarshal([]byte(buffer), &result)
	if err != nil {
		return result, fmt.Errorf("failed to unmarshal productData: %w", err)
	}
	return result, nil
}

func (w *Woolworths) Init(baseURL string, dbPath string) {
	var err error
	w.cookieJar, err = cookiejar.New(nil)
	if err != nil {
		slog.Error(fmt.Sprintf("Error creating cookie jar: %v", err))
	}
	w.baseURL = baseURL
	w.client = &RLHTTPClient{
		client: &http.Client{
			Jar: w.cookieJar,
		},
		Ratelimiter: rate.NewLimiter(rate.Every(1*time.Second), 1),
	}
	w.InitDB(dbPath)
}

// This produces a stream of product IDs that are expired and need an update.
func (w *Woolworths) ProductUpdateQueueWorker(output chan ProductID, maxAge time.Duration) {
	for {
		var productID ProductID
		var t time.Time
		err := w.db.QueryRow(`	SELECT productID, retrieved FROM products
								WHERE retrieved < ?
								ORDER BY retrieved ASC
								LIMIT 1`, time.Now().Add(-maxAge)).Scan(&productID, &t)
		if err != nil {
			if err != sql.ErrNoRows {
				slog.Error(fmt.Sprintf("Error getting product ID: %v", err))
			}
		} else {
			slog.Debug("Retrieved time", "time", t)
			slog.Debug("current time", "time", time.Now())
			slog.Debug("threshold time", "time", time.Now().Add(-maxAge))
			slog.Debug("ProductUpdateQueueWorker Blocking on sending out")
			output <- productID
			slog.Debug("ProductUpdateQueueWorker Unblocked on sending out")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func (w *Woolworths) NewProductIDWorker(output chan ProductID) {
	// TODO
	output <- 165262
	output <- 187314
	output <- 524336
}

func (w *Woolworths) RunScheduler() {

	slog.SetLogLoggerLevel(slog.LevelDebug)

	// productIDInputChannel := make(chan ProductID)
	productInfoChannel := make(chan WoolworthsProductInfo)
	newProductIDsChannel := make(chan ProductID)
	productsThatNeedAnUpdateChannel := make(chan ProductID)
	newDepartmentIDsChannel := make(chan ProductID)
	go w.ProductWorker(productsThatNeedAnUpdateChannel, productInfoChannel)
	go w.ProductUpdateQueueWorker(productsThatNeedAnUpdateChannel, PRODUCT_INFO_MAX_AGE_SECONDS)
	// go w.ProductUpdateQueueWorker(productsThatNeedAnUpdateChannel, 1000*time.Millisecond)
	go w.NewProductIDWorker(newProductIDsChannel)

	for {
		slog.Debug("Loopan")
		select {
		// case productID := <-productsThatNeedAnUpdateChannel:
		// 	slog.Debug(fmt.Sprintf("Product ID that needs an update: %d", productID))
		// 	productIDInputChannel <- productID
		case productInfoUpdate := <-productInfoChannel:
			slog.Debug(fmt.Sprintf("Product info update: %v", productInfoUpdate))
			// Update the product info in the DB
			err := w.SaveProductInfo(productInfoUpdate, time.Now())
			if err != nil {
				slog.Error(fmt.Sprintf("Error saving product info: %v", err))
			}
		case newProductID := <-newProductIDsChannel:
			slog.Debug(fmt.Sprintf("New product ID: %d", newProductID))
			// Update the productIDs table with the new product ID
			// with blank productinfo and and a very old datetime.
			err := w.SaveProductInfo(WoolworthsProductInfo{ID: newProductID, Info: ProductInfo{}}, time.Now().Add(-PRODUCT_INFO_MAX_AGE_SECONDS))
			if err != nil {
				slog.Error(fmt.Sprintf("Error saving product info: %v", err))
			}
		case newDepartmentID := <-newDepartmentIDsChannel:
			slog.Debug(fmt.Sprintf("New department ID: %d", newDepartmentID))
			// Update the departmentIDs table with the new department ID
		}
	}

}
