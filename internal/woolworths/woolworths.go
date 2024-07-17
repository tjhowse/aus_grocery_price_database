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

type Woolworths struct {
	baseURL   string
	client    *RLHTTPClient
	cookieJar *cookiejar.Jar // TODO This might not be threadsafe.
	db        *sql.DB
}

func (w *Woolworths) ProductWorker(input chan ProductID, output chan ProductInfo) {
	slog.Debug("Running a Woolworths.Worker")
	for id := range input {
		slog.Debug(fmt.Sprintf("Getting product info for ID: %d", id))
		info, err := w.GetProductInfo(id)
		if err != nil {
			slog.Error(fmt.Sprintf("Error getting product info: %v", err))
		}
		output <- info
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
func (w *Woolworths) SaveProductInfo(productInfo ProductInfo) error {
	var err error
	var result sql.Result

	var productInfoBytes []byte

	productInfoBytes, err = json.Marshal(productInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal product info: %w", err)
	}
	productInfoString := string(productInfoBytes)

	result, err = w.db.Exec(`
		INSERT INTO products (productID, productData, retrieved)
		VALUES (?, ?, ?)
		ON CONFLICT(productID) DO UPDATE SET productID = ?, productData = ?, retrieved = ?
		`, productInfo.ProductID, productInfoString, time.Now(), productInfo.ProductID, productInfoString, time.Now())

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
	var result ProductInfo
	err := w.db.QueryRow("SELECT productData FROM products WHERE productID = ? LIMIT 1", productID).Scan(&result)
	if err != nil {
		return result, fmt.Errorf("failed to query existing productData: %w", err)
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

func (w *Woolworths) RunScheduler() {

	productIDInputChannel := make(chan ProductID)
	productInfoChannel := make(chan ProductInfo)
	newProductIDsChannel := make(chan ProductID)
	newDepartmentIDsChannel := make(chan ProductID)
	go w.ProductWorker(productIDInputChannel, productInfoChannel)

	select {
	case newProductID := <-newProductIDsChannel:
		slog.Debug(fmt.Sprintf("New product ID: %d", newProductID))
		// Update the productIDs table with the new product ID
	case newDepartmentID := <-newDepartmentIDsChannel:
		slog.Debug(fmt.Sprintf("New department ID: %d", newDepartmentID))
		// Update the departmentIDs table with the new department ID
	case productInfo := <-productInfoChannel:
		slog.Debug(fmt.Sprintf("Product info: %v", productInfo))
		// Update the products table with the new product info
	}

}
