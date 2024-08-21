package woolworths

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/shopspring/decimal"
)

const DB_SCHEMA_VERSION = 6

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
	_, err = w.db.Exec("CREATE TABLE IF NOT EXISTS departments (departmentID TEXT UNIQUE, description TEXT, productCount INTEGER, updated DATETIME)")
	if err != nil {
		return err
	}
	_, err =
		w.db.Exec(`	CREATE TABLE IF NOT EXISTS products
						(	productID TEXT UNIQUE,
							name TEXT,
							description TEXT,
							barcode TEXT,
							priceCents INTEGER,
							weightGrams INTEGER,
							productJSON TEXT,
							departmentID TEXT DEFAULT "",
							updated DATETIME
						)`)
	if err != nil {
		return err
	}
	return nil
}

// backupDB moves the specified DB to the same directory with an ISO8601 timestamp and the schema
// number prepended to the filename.
func (w *Woolworths) backupDB(dbPath string, oldSchema int) error {
	backupName := fmt.Sprintf("%s.%d.%s", dbPath, oldSchema, time.Now().Format("2006-01-02T15:04:05"))
	err := os.Rename(dbPath, backupName)
	if err != nil {
		return fmt.Errorf("failed to backup existing DB: %w", err)
	}
	slog.Info("Backed up old DB", "old", dbPath, "new", backupName)
	return nil
}

func openDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath+"?cache=shared")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	return db, nil
}

func (w *Woolworths) initDB(dbPath string) error {
	var err error
	w.db, err = openDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open DB: %w", err)
	}
	var version int
	err = w.db.QueryRow("SELECT version FROM schema").Scan(&version)

	if err != nil || version != DB_SCHEMA_VERSION {
		slog.Warn("DB schema mismatch", "path", dbPath, "currentVersion", DB_SCHEMA_VERSION, "detectedVersion", version)

		if version != 0 {
			// If we detected an old schema, backup the DB and create a new one.
			err = w.db.Close()
			if err != nil {
				return fmt.Errorf("failed to close existing DB before backing it up: %w", err)
			}
			err = w.backupDB(dbPath, version)
			if err != nil {
				return fmt.Errorf("failed to backup existing DB: %w", err)
			}

			// Open a new DB
			w.db, err = openDB(dbPath)
			if err != nil {
				return fmt.Errorf("failed to open DB: %w", err)
			}
		}

		// Create the schema
		err := w.initBlankDB()
		if err != nil {
			return fmt.Errorf("failed to create blank DB: %w", err)
		} else {
			slog.Info("New blank DB created")
		}
	}
	return nil
}

// Saves product info to the database
func (w *Woolworths) saveProductInfoExtended(tx *sql.Tx, productInfo woolworthsProductInfoExtended) error {
	var err error
	var result sql.Result

	result, err = tx.Exec(`
			INSERT INTO products (productID, name, description, barcode, priceCents, weightGrams, productJSON, departmentID, updated)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(productID) DO UPDATE SET
				productID = excluded.productID,
				name = excluded.name,
				description = excluded.description,
				barcode = excluded.barcode,
				priceCents = excluded.priceCents,
				weightGrams = excluded.weightGrams,
				productJSON = excluded.productJSON,
				departmentID = excluded.departmentID,
				updated = excluded.updated`,
		productInfo.ID, productInfo.Info.DisplayName, productInfo.Info.Description, productInfo.Info.Barcode,
		productInfo.Info.Price.Mul(decimal.NewFromInt(100)).IntPart(),
		productInfo.Info.UnitWeightInGrams, productInfo.RawJSON, productInfo.departmentID, productInfo.Updated)

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
func (w *Woolworths) saveProductInfoExtendedNoTx(productInfo woolworthsProductInfoExtended) error {
	var err error

	tx, err := w.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	w.saveProductInfoExtended(tx, productInfo)
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// Saves product info to the database
func (w *Woolworths) saveDepartment(departmentInfo departmentInfo) error {
	var err error
	var result sql.Result

	result, err = w.db.Exec(`
		INSERT INTO departments (departmentID, description, productCount, updated)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(departmentID) DO UPDATE SET
			departmentID = excluded.departmentID,
			description = excluded.description,
			productCount = excluded.productCount,
			updated = excluded.updated`,
		departmentInfo.NodeID, departmentInfo.Description, departmentInfo.ProductCount, departmentInfo.Updated)

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

// loadProductInfoExtended loads cached extended product info from the database
func (w *Woolworths) loadProductInfoExtended(productID productID) (woolworthsProductInfoExtended, error) {
	var wProdInfo woolworthsProductInfoExtended
	var deptDescription sql.NullString
	row := w.db.QueryRow(`
	SELECT
		productID,
		name,
		products.description,
		barcode,
		priceCents,
		weightGrams,
		productJSON,
		products.departmentID,
		departments.description,
		products.updated
	FROM
		products
		LEFT JOIN departments ON products.departmentID = departments.departmentID
	WHERE productID = ? LIMIT 1`, productID)
	err := row.Scan(
		&wProdInfo.ID,
		&wProdInfo.Info.DisplayName,
		&wProdInfo.Info.Description,
		&wProdInfo.Info.Barcode,
		&wProdInfo.Info.Price,
		&wProdInfo.Info.UnitWeightInGrams,
		&wProdInfo.RawJSON,
		&wProdInfo.departmentID,
		&deptDescription, // This value comes from a join, so it might be NULL.
		&wProdInfo.Updated)
	if err != nil {
		if err == sql.ErrNoRows {
			return wProdInfo, ErrProductMissing
		}
		return wProdInfo, fmt.Errorf("failed to query existing product info: %w", err)
	}
	if deptDescription.Valid {
		wProdInfo.departmentDescription = deptDescription.String
	}
	return wProdInfo, nil
}

func (w *Woolworths) loadDepartmentInfoList() ([]departmentInfo, error) {
	var departmentInfos []departmentInfo
	rows, err := w.db.Query("SELECT departmentID, description, productCount, updated FROM departments")
	if err != nil {
		return departmentInfos, fmt.Errorf("failed to query departmentIDs: %w", err)
	}
	for rows.Next() {
		var deptInfo departmentInfo
		err = rows.Scan(&deptInfo.NodeID, &deptInfo.Description, &deptInfo.ProductCount, &deptInfo.Updated)
		if err != nil {
			return departmentInfos, fmt.Errorf("failed to scan departmentID: %w", err)
		}
		departmentInfos = append(departmentInfos, deptInfo)
	}
	return departmentInfos, nil
}
