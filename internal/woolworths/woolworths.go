package woolworths

import (
	"database/sql"
	"encoding/json"
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

func (w *Woolworths) ProductInfoFetchingWorker(input chan ProductID, output chan WoolworthsProductInfo) {
	slog.Debug("Running a Woolworths.Worker")
	for id := range input {
		slog.Debug(fmt.Sprintf("Getting product info for ID: %d", id))
		info, err := w.GetProductInfo(id)
		if err != nil {
			slog.Error(fmt.Sprintf("Error getting product info: %v", err))
		}
		info.Updated = time.Now()
		output <- info
	}
}

// Initialises the DB with the schema. Note you must bump the DB_SCHEMA_VERSION
// constant if you change the schema.
func (w *Woolworths) InitBlankDB() {
	_, err := w.db.Exec("CREATE TABLE IF NOT EXISTS schema (version INTEGER PRIMARY KEY)")
	if err != nil {
		slog.Error(fmt.Sprintf("Error creating schema table: %v", err))
	}
	_, err = w.db.Exec("INSERT INTO schema (version) VALUES (?)", DB_SCHEMA_VERSION)
	w.db.Exec("CREATE TABLE IF NOT EXISTS departmentIDs (departmentID TEXT UNIQUE, updated DATETIME)")
	if err != nil {
		slog.Error(fmt.Sprintf("Error creating schema table: %v", err))
	}
	_, err =
		w.db.Exec("CREATE TABLE IF NOT EXISTS productIDs (productID INTEGER UNIQUE, updated DATETIME)")
	if err != nil {
		slog.Error(fmt.Sprintf("Error creating schema table: %v", err))
	}
	_, err =
		w.db.Exec("CREATE TABLE IF NOT EXISTS products (productID INTEGER UNIQUE, productData TEXT, updated DATETIME)")
	if err != nil {
		slog.Error(fmt.Sprintf("Error creating schema table: %v", err))
	}
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

		Each row in each table has a updated datetime field.
		In code, define a maximum age for each type of data.
		The runner will check the age of the data in the DB and refresh it if it's too old.
		If a new product is found, it will be added to the DB and have its datatime set to 0.
		The scheduler queries the DB sorted by datetime and launches a worker to refresh the data
		as required. The workers don't write to the DB, they just return the data to the scheduler.
	*/
}

// Saves product info to the database
func (w *Woolworths) SaveProductInfo(productInfo WoolworthsProductInfo) error {
	var err error
	var result sql.Result

	var productInfoBytes []byte

	slog.Debug("Saving product", "name", productInfo.Info.Name)

	productInfoBytes, err = json.Marshal(productInfo.Info)
	if err != nil {
		return fmt.Errorf("failed to marshal product info: %w", err)
	}
	productInfoString := string(productInfoBytes)

	result, err = w.db.Exec(`
		INSERT INTO products (productID, productData, updated)
		VALUES (?, ?, ?)
		ON CONFLICT(productID) DO UPDATE SET productID = ?, productData = ?, updated = ?
		`, productInfo.ID, productInfoString, productInfo.Updated, productInfo.ID, productInfoString, productInfo.Updated)

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
		if err == sql.ErrNoRows {
			return result, ErrProductMissing
		}
		return result, fmt.Errorf("failed to query existing productData: %w", err)
	}
	err = json.Unmarshal([]byte(buffer), &result)
	if err != nil {
		return result, fmt.Errorf("failed to unmarshal productData: %w", err)
	}
	return result, nil
}

func (w *Woolworths) LoadDepartmentIDsList() ([]DepartmentID, error) {
	var departmentIDs []DepartmentID
	rows, err := w.db.Query("SELECT departmentID FROM departmentIDs")
	if err != nil {
		return departmentIDs, fmt.Errorf("failed to query departmentIDs: %w", err)
	}
	for rows.Next() {
		var departmentID DepartmentID
		err = rows.Scan(&departmentID)
		if err != nil {
			return departmentIDs, fmt.Errorf("failed to scan departmentID: %w", err)
		}
		departmentIDs = append(departmentIDs, departmentID)
	}
	return departmentIDs, nil
}

func (w *Woolworths) Init(baseURL string, dbPath string, productMaxAge time.Duration) {
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
	w.productMaxAge = productMaxAge
	w.InitDB(dbPath)
}

// This produces a stream of product IDs that are expired and need an update.
func (w *Woolworths) ProductUpdateQueueWorker(output chan ProductID, maxAge time.Duration) {
	for {
		var productIDs []ProductID
		rows, err := w.db.Query(`	SELECT productID FROM products
									WHERE updated < ?
									ORDER BY updated ASC
									LIMIT 10`, time.Now().Add(-maxAge))
		if err != nil {
			if err != sql.ErrNoRows {
				slog.Error(fmt.Sprintf("Error getting product ID: %v", err))
			}
		} else {
			for rows.Next() {
				var productID ProductID
				err = rows.Scan(&productID)
				if err != nil {
					slog.Error(fmt.Sprintf("Error scanning product ID: %v", err))
				}
				slog.Debug("Product ID needs an update", "productID", productID)
				productIDs = append(productIDs, productID)
			}
		}

		for _, productID := range productIDs {
			output <- productID
		}
		time.Sleep(1 * time.Second)
	}
}

func (w *Woolworths) NewDepartmentIDWorker(output chan DepartmentID) {
	for {
		// Read the department list from the web...
		departmentsFromWeb, err := w.GetDepartmentIDs()
		if err != nil {
			slog.Error(fmt.Sprintf("Error getting department IDs from web: %v", err))
		}

		// Read the department list from the DB.
		departmentsFromDB, err := w.LoadDepartmentIDsList()
		if err != nil {
			slog.Error(fmt.Sprintf("Error loading department IDs from DB: %v", err))
		}
		// Compare the two lists and output any new department IDs.
		for _, webDepartmentID := range departmentsFromWeb {
			found := false
			for _, dbDepartmentID := range departmentsFromDB {
				if webDepartmentID == dbDepartmentID {
					found = true
					break
				}
			}
			if found {
				continue
			}
			output <- webDepartmentID
		}
		// We don't need to check for departments very often.
		time.Sleep(1 * time.Hour)
	}
}

func (w *Woolworths) NewProductIDWorker(output chan WoolworthsProductInfo) {
	// TODO
	// output <- WoolworthsProductInfo{ID: 165262, Info: ProductInfo{}, Updated: time.Now().Add(-2 * w.productMaxAge)}
	// output <- WoolworthsProductInfo{ID: 187314, Info: ProductInfo{}, Updated: time.Now().Add(-2 * w.productMaxAge)}
	// output <- WoolworthsProductInfo{ID: 524336, Info: ProductInfo{}, Updated: time.Now().Add(-2 * w.productMaxAge)}
	// return
	for {
		departments, err := w.LoadDepartmentIDsList()
		if err != nil {
			slog.Error(fmt.Sprintf("Error loading department IDs: %v", err))
			// Try again in ten minutes.
			time.Sleep(10 * time.Minute)
			continue
		}
		for _, departmentID := range departments {
			products, err := w.GetProductsFromDepartment(departmentID)
			if err != nil {
				slog.Error(fmt.Sprintf("Error getting products from department: %v", err))
				// Try again in ten minutes.
				time.Sleep(10 * time.Minute)
				continue
			}
			for _, productID := range products {
				_, err := w.LoadProductInfo(productID)
				if err != ErrProductMissing {
					continue
				}
				output <- WoolworthsProductInfo{ID: productID, Info: ProductInfo{}, Updated: time.Now().Add(-2 * w.productMaxAge)}
			}
		}
		// We don't need to check for new products very often.
		time.Sleep(1 * time.Hour)
	}
}

// Runs up all the workers and mediates data flowing between them.
// Currently all sqlite writes happen via this function. This may move
// off to a separate goroutine in the future.
func (w *Woolworths) RunScheduler(cancel chan struct{}) {

	productInfoChannel := make(chan WoolworthsProductInfo)
	productsThatNeedAnUpdateChannel := make(chan ProductID)
	newDepartmentIDsChannel := make(chan DepartmentID)
	// for i := 0; i < PRODUCT_INFO_WORKER_COUNT; i++ {
	// 	go w.ProductInfoFetchingWorker(productsThatNeedAnUpdateChannel, productInfoChannel)
	// }
	go w.ProductInfoFetchingWorker(productsThatNeedAnUpdateChannel, productInfoChannel)
	go w.ProductUpdateQueueWorker(productsThatNeedAnUpdateChannel, w.productMaxAge)
	go w.NewProductIDWorker(productInfoChannel)
	go w.NewDepartmentIDWorker(newDepartmentIDsChannel)

	for {
		slog.Debug("Loopan")
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
		case <-cancel:
			slog.Info("Exiting scheduler")
			return
		}
	}

}
