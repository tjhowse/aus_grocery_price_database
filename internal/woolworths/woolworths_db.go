package woolworths

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/tjhowse/aus_grocery_price_database/internal/shared"
)

const DB_SCHEMA_VERSION = 2

// Initialises the DB with the schema. Note you must bump the DB_SCHEMA_VERSION
// constant if you change the schema.
func (w *Woolworths) initBlankDB() error {

	// Drop all tables
	for _, table := range []string{"schema", "departmentIDs", "products"} {
		// Mildly confused by why this doesn't work? TODO investigate
		// _, err := w.db.Exec("DROP TABLE IF EXISTS ?", table)
		_, err := w.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table))
		if err != nil {
			return err
		}
	}

	_, err := w.db.Exec("CREATE TABLE IF NOT EXISTS schema (version INTEGER PRIMARY KEY)")
	if err != nil {
		return err
	}
	_, err = w.db.Exec("INSERT INTO schema (version) VALUES (?)", DB_SCHEMA_VERSION)
	if err != nil {
		return err
	}
	_, err = w.db.Exec("CREATE TABLE IF NOT EXISTS departmentIDs (departmentID TEXT UNIQUE, updated DATETIME)")
	if err != nil {
		return err
	}
	_, err =
		w.db.Exec(`	CREATE TABLE IF NOT EXISTS products
						(	productID TEXT UNIQUE,
							name TEXT,
							description TEXT,
							price FLOAT,
							weightGrams FLOAT,
							productJSON TEXT,
							updated DATETIME
						)`)
	if err != nil {
		return err
	}
	return nil
}

func (w *Woolworths) initDB(dbPath string) error {
	var err error
	dbPath += "?cache=shared"
	w.db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	w.db.SetMaxOpenConns(1)
	var version int
	err = w.db.QueryRow("SELECT version FROM schema").Scan(&version)
	if err != nil {
		slog.Warn("DB schema error. Creating blank DB.", "error", err)
	}
	if version != DB_SCHEMA_VERSION {
		slog.Warn("DB schema error. Creating blank DB.", "path", dbPath, "version", DB_SCHEMA_VERSION)
		err := w.initBlankDB()
		if err != nil {
			return fmt.Errorf("failed to create blank DB: %w", err)
		} else {
			slog.Info("Blank DB created")
		}
	}
	return nil
}

// Saves product info to the database
func (w *Woolworths) saveProductInfo(productInfo woolworthsProductInfo) error {
	var err error
	var result sql.Result

	productInfoString := string(productInfo.RawJSON)

	// TODO Is there some better way of handling passing copies of the same data?
	result, err = w.db.Exec(`
		INSERT INTO products (productID, name, description, price, weightGrams, productJSON, updated)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(productID) DO UPDATE SET productID = ?, name = ?, description = ?, price = ?, weightGrams = ?, productJSON = ?, updated = ?
		`, productInfo.ID, productInfo.Info.Name, productInfo.Info.Description,
		productInfo.Info.Offers.Price, productInfo.Info.Weight, productInfoString, productInfo.Updated,
		productInfo.ID, productInfo.Info.Name, productInfo.Info.Description,
		productInfo.Info.Offers.Price, productInfo.Info.Weight, productInfoString, productInfo.Updated)

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

// Saves product info to the database
func (w *Woolworths) saveDepartment(productInfo departmentID) error {
	var err error
	var result sql.Result

	result, err = w.db.Exec(`
		INSERT INTO departmentIDs (departmentID, updated)
		VALUES (?, ?)
		ON CONFLICT(departmentID) DO UPDATE SET departmentID = ?, updated = ?
		`, productInfo, time.Now(), productInfo, time.Now())

	if err != nil {
		return fmt.Errorf("failed to update department ID info: %w", err)
	}
	if rowsAffected, err := result.RowsAffected(); err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	} else if rowsAffected == 0 {
		slog.Warn("Department not upserted")
	}

	return nil
}

// Loads cached product info from the database
func (w *Woolworths) loadProductInfo(productID productID) (productInfo, error) {
	var buffer string
	var result productInfo
	err := w.db.QueryRow("SELECT productJSON FROM products WHERE productID = ? LIMIT 1", productID).Scan(&buffer)
	if err != nil {
		if err == sql.ErrNoRows {
			return result, ErrProductMissing
		}
		return result, fmt.Errorf("failed to query existing productJSON: %w", err)
	}
	err = json.Unmarshal([]byte(buffer), &result)
	if err != nil {
		return result, fmt.Errorf("failed to unmarshal productJSON: %w", err)
	}
	return result, nil
}

func (w *Woolworths) loadDepartmentIDsList() ([]departmentID, error) {
	var departmentIDs []departmentID
	rows, err := w.db.Query("SELECT departmentID FROM departmentIDs")
	if err != nil {
		return departmentIDs, fmt.Errorf("failed to query departmentIDs: %w", err)
	}
	for rows.Next() {
		var departmentID departmentID
		err = rows.Scan(&departmentID)
		if err != nil {
			return departmentIDs, fmt.Errorf("failed to scan departmentID: %w", err)
		}
		departmentIDs = append(departmentIDs, departmentID)
	}
	return departmentIDs, nil
}

// Returns a list of product IDs that have been updated since the given time
func (w *Woolworths) GetSharedProductsUpdatedAfter(t time.Time, count int) ([]shared.ProductInfo, error) {
	var productIDs []shared.ProductInfo
	rows, err := w.db.Query("SELECT productID, name, description, price, weightGrams, updated FROM products WHERE updated > ? LIMIT ?", t, count)
	if err != nil {
		return productIDs, fmt.Errorf("failed to query productIDs: %w", err)
	}
	for rows.Next() {
		var product shared.ProductInfo
		err = rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.WeightGrams, &product.Timestamp)
		if err != nil {
			return productIDs, fmt.Errorf("failed to scan productID: %w", err)
		}
		product.ID = WOOLWORTHS_ID_PREFIX + product.ID
		productIDs = append(productIDs, product)
	}
	return productIDs, nil
}
