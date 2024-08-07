package woolworths

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
)

const DB_SCHEMA_VERSION = 4

// Initialises the DB with the schema. Note you must bump the DB_SCHEMA_VERSION
// constant if you change the schema.
func (w *Woolworths) initBlankDB() error {

	// Drop all tables
	for _, table := range []string{"schema", "departments", "products"} {
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
	_, err = w.db.Exec("CREATE TABLE IF NOT EXISTS departments (departmentID TEXT UNIQUE, description TEXT, updated DATETIME)")
	if err != nil {
		return err
	}
	_, err =
		w.db.Exec(`	CREATE TABLE IF NOT EXISTS products
						(	productID TEXT UNIQUE,
							name TEXT,
							description TEXT,
							priceCents INTEGER,
							weightGrams INTEGER,
							productJSON TEXT,
							departmentID TEXT DEFAULT "",
							departmentDescription TEXT DEFAULT "",
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
	// TODO I bet a real SQL wizard could combine these two statements such that the department
	// 		info is only written to the DB if it is not empty in the productInfo struct.
	if productInfo.departmentID != "" {
		// If we have department info, this data must've come from a department update.
		// Only save the department info and leave the rest alone.

		result, err = w.db.Exec(`
			INSERT INTO products (productID, departmentID, departmentDescription, updated)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(productID) DO UPDATE SET productID = ?, departmentID = ?, departmentDescription = ?, updated = ?`,
			productInfo.ID, productInfo.departmentID, productInfo.departmentDescription, productInfo.Updated,
			productInfo.ID, productInfo.departmentID, productInfo.departmentDescription, productInfo.Updated)

		if err != nil {
			return fmt.Errorf("failed to update product info: %w", err)
		}
		if rowsAffected, err := result.RowsAffected(); err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		} else if rowsAffected == 0 {
			slog.Warn("Product department info not updated.")
		}

	}

	result, err = w.db.Exec(`
			INSERT INTO products (productID, name, description, priceCents, weightGrams, productJSON, updated)
			VALUES (?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(productID) DO UPDATE SET productID = ?, name = ?, description = ?, priceCents = ?, weightGrams = ?, productJSON = ?, updated = ?`,
		productInfo.ID, productInfo.Info.Name, productInfo.Info.Description,
		productInfo.Info.Offers.Price.Mul(decimal.NewFromInt(100)).IntPart(),
		productInfo.Info.Weight, productInfoString, productInfo.Updated,

		productInfo.ID, productInfo.Info.Name, productInfo.Info.Description,
		productInfo.Info.Offers.Price.Mul(decimal.NewFromInt(100)).IntPart(),
		productInfo.Info.Weight, productInfoString, productInfo.Updated)

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
func (w *Woolworths) saveProductInfoExtended(productInfo woolworthsProductInfoExtended) error {
	var err error
	var result sql.Result

	productInfoString := "raw json todo"

	result, err = w.db.Exec(`
			INSERT INTO products (productID, name, description, priceCents, weightGrams, productJSON, updated)
			VALUES (?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(productID) DO UPDATE SET productID = ?, name = ?, description = ?, priceCents = ?, weightGrams = ?, productJSON = ?, updated = ?`,
		productInfo.ID, productInfo.Info.Name, productInfo.Info.Description,
		productInfo.Info.Price.Mul(decimal.NewFromInt(100)).IntPart(),
		productInfo.Info.UnitWeightInGrams, productInfoString, productInfo.Updated,

		productInfo.ID, productInfo.Info.Name, productInfo.Info.Description,
		productInfo.Info.Price.Mul(decimal.NewFromInt(100)).IntPart(),
		productInfo.Info.UnitWeightInGrams, productInfoString, productInfo.Updated)

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
func (w *Woolworths) saveDepartment(departmentInfo departmentInfo) error {
	var err error
	var result sql.Result

	result, err = w.db.Exec(`
		INSERT INTO departments (departmentID, description, updated)
		VALUES (?, ?, ?)
		ON CONFLICT(departmentID) DO UPDATE SET departmentID = ?, description = ?, updated = ?
		`,
		departmentInfo.NodeID, departmentInfo.Description, time.Now(),
		departmentInfo.NodeID, departmentInfo.Description, time.Now())

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
func (w *Woolworths) loadProductInfo(productID productID) (woolworthsProductInfo, error) {
	var wProdInfo woolworthsProductInfo
	row := w.db.QueryRow(`SELECT
		productID, name, description, priceCents, weightGrams, departmentID, departmentDescription, updated
		FROM products WHERE productID = ? LIMIT 1`, productID)
	err := row.Scan(
		&wProdInfo.ID,
		&wProdInfo.Info.Name,
		&wProdInfo.Info.Description,
		&wProdInfo.Info.Offers.Price,
		&wProdInfo.Info.Weight,
		&wProdInfo.departmentID,
		&wProdInfo.departmentDescription,
		&wProdInfo.Updated)
	if err != nil {
		if err == sql.ErrNoRows {
			return wProdInfo, ErrProductMissing
		}
		return wProdInfo, fmt.Errorf("failed to query existing product info: %w", err)
	}

	return wProdInfo, nil
}

// Returns true if the product ID exists in the DB already.
// This exists separately to loadProductInfo because the productJSON is quite big.
// This should be faster.
func (w *Woolworths) checkIfKnownProductID(productID productID) (bool, error) {
	var count int
	err := w.db.QueryRow("SELECT COUNT(productID) FROM products WHERE productID = ? LIMIT 1", productID).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check for existing product: %w", err)
	}
	return count > 0, nil
}

func (w *Woolworths) loadDepartmentInfoList() ([]departmentInfo, error) {
	var departmentInfos []departmentInfo
	rows, err := w.db.Query("SELECT departmentID, description FROM departments")
	if err != nil {
		return departmentInfos, fmt.Errorf("failed to query departmentIDs: %w", err)
	}
	for rows.Next() {
		var deptInfo departmentInfo
		err = rows.Scan(&deptInfo.NodeID, &deptInfo.Description)
		if err != nil {
			return departmentInfos, fmt.Errorf("failed to scan departmentID: %w", err)
		}
		departmentInfos = append(departmentInfos, deptInfo)
	}
	return departmentInfos, nil
}
