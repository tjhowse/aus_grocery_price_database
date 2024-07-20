package woolworths

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

// Initialises the DB with the schema. Note you must bump the DB_SCHEMA_VERSION
// constant if you change the schema.
func (w *Woolworths) InitBlankDB() error {

	// Drop all tables
	for _, table := range []string{"schema", "departmentIDs", "productIDs", "products"} {
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
		w.db.Exec("CREATE TABLE IF NOT EXISTS productIDs (productID INTEGER UNIQUE, updated DATETIME)")
	if err != nil {
		return err
	}
	_, err =
		w.db.Exec("CREATE TABLE IF NOT EXISTS products (productID INTEGER UNIQUE, productData TEXT, updated DATETIME)")
	if err != nil {
		return err
	}
	return nil
}

func (w *Woolworths) InitDB(dbPath string) error {
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
		err := w.InitBlankDB()
		if err != nil {
			return fmt.Errorf("failed to create blank DB: %w", err)
		} else {
			slog.Info("Blank DB created")
		}
	}
	return nil
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

// Saves product info to the database
func (w *Woolworths) SaveDepartment(productInfo DepartmentID) error {
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

// Returns a list of product IDs that have been updated since the given time
func (w *Woolworths) GetProductIDsUpdatedAfter(t time.Time) ([]ProductID, error) {
	var productIDs []ProductID
	rows, err := w.db.Query("SELECT productID FROM products WHERE updated > ?", t)
	if err != nil {
		return productIDs, fmt.Errorf("failed to query productIDs: %w", err)
	}
	for rows.Next() {
		var productID ProductID
		err = rows.Scan(&productID)
		if err != nil {
			return productIDs, fmt.Errorf("failed to scan productID: %w", err)
		}
		productIDs = append(productIDs, productID)
	}
	return productIDs, nil
}
