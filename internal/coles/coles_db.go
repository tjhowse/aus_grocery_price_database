package coles

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/shopspring/decimal"
	"github.com/tjhowse/aus_grocery_price_database/internal/shared"
)

const DB_SCHEMA_VERSION = 1

// Initialises the DB with the schema. Note you must bump the DB_SCHEMA_VERSION
// constant if you change the schema.
func (w *Coles) initBlankDB() error {

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
							previousPriceCents INTEGER,
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
func (w *Coles) backupDB(dbPath string, oldSchema int) error {
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

// initDB initialises the database.
func (c *Coles) initDB(dbPath string) error {
	var err error
	c.db, err = openDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open DB: %w", err)
	}
	var version int
	err = c.db.QueryRow("SELECT version FROM schema").Scan(&version)

	if err != nil || version != DB_SCHEMA_VERSION {
		slog.Warn("DB schema mismatch", "path", dbPath, "currentVersion", DB_SCHEMA_VERSION, "detectedVersion", version)

		if version != 0 {
			// If we detected an old schema, backup the DB and create a new one.
			err = c.db.Close()
			if err != nil {
				return fmt.Errorf("failed to close existing DB before backing it up: %w", err)
			}
			err = c.backupDB(dbPath, version)
			if err != nil {
				return fmt.Errorf("failed to backup existing DB: %w", err)
			}

			// Open a new DB
			c.db, err = openDB(dbPath)
			if err != nil {
				return fmt.Errorf("failed to open DB: %w", err)
			}
		}

		// Create the schema
		err := c.initBlankDB()
		if err != nil {
			return fmt.Errorf("failed to create blank DB: %w", err)
		} else {
			slog.Info("New blank DB created")
		}
	}
	return nil
}

// saveProductInfo saves the product info to the database transactionfully.
func (c *Coles) saveProductInfoes(products []colesProductInfo) error {
	tx, err := c.db.Begin()
	defer tx.Rollback()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	for _, product := range products {
		if err := c.saveProductInfo(tx, product); err != nil {
			return fmt.Errorf("failed to save product info: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func calcWeightInGrams(productInfo colesProductInfo) (int, error) {
	scalar := 0.0
	switch productInfo.Info.Pricing.Unit.OfMeasureUnits {
	case "g":
		scalar = 1.0
	case "kg":
		scalar = 1000.0
	default:
		return 0, fmt.Errorf("cannot convert unit `%s` to grams", productInfo.Info.Pricing.Unit.OfMeasureUnits)
	}
	return int(float64(productInfo.Info.Pricing.Unit.Quantity) * scalar), nil
}

// saveProductInfo saves a single product to the database transactionfully.
func (c *Coles) saveProductInfo(tx *sql.Tx, productInfo colesProductInfo) error {
	var err error
	var result sql.Result

	productInfo.WeightGrams, err = calcWeightInGrams(productInfo)
	if err != nil {
		slog.Warn("Couldn't calculate weight in grams", "productID", productInfo.ID, "error", err)
		productInfo.WeightGrams = 0
	}

	result, err = tx.Exec(`
			INSERT INTO products (productID, name, description, barcode, priceCents, previousPriceCents, weightGrams, productJSON, departmentID, updated)
			VALUES (?, ?, ?, ?, ?, 0, ?, ?, ?, ?)
			ON CONFLICT(productID) DO UPDATE SET
				productID = excluded.productID,
				name = excluded.name,
				description = excluded.description,
				barcode = excluded.barcode,
				priceCents = excluded.priceCents,
				previousPriceCents = priceCents,
				weightGrams = excluded.weightGrams,
				productJSON = excluded.productJSON,
				departmentID = excluded.departmentID,
				updated = excluded.updated`,
		productInfo.ID, productInfo.Info.Name, productInfo.Info.Description, 0,
		productInfo.Info.Pricing.Now.Mul(decimal.NewFromInt(100)).IntPart(),
		productInfo.WeightGrams, productInfo.RawJSON, productInfo.departmentID, productInfo.Updated)

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

// loadProductInfo loads cached extended product info from the database
func (w *Coles) loadProductInfo(productID productID) (colesProductInfo, error) {
	var cProdInfo colesProductInfo
	var deptDescription sql.NullString
	row := w.db.QueryRow(`
	SELECT
		productID,
		name,
		products.description,
		priceCents,
		previousPriceCents,
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
		&cProdInfo.ID,
		&cProdInfo.Info.Name,
		&cProdInfo.Info.Description,
		&cProdInfo.Info.Pricing.Now,
		&cProdInfo.PreviousPrice,
		&cProdInfo.WeightGrams,
		&cProdInfo.RawJSON,
		&cProdInfo.departmentID,
		&deptDescription, // This value comes from a join, so it might be NULL.
		&cProdInfo.Updated)
	if err != nil {
		if err == sql.ErrNoRows {
			return cProdInfo, shared.ErrProductMissing
		}
		return cProdInfo, fmt.Errorf("failed to query existing product info: %w", err)
	}
	if deptDescription.Valid {
		cProdInfo.departmentDescription = deptDescription.String
	}
	return cProdInfo, nil
}
