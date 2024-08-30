package coles

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
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