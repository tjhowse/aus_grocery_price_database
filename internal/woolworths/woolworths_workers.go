package woolworths

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

// This produces a stream of product IDs that are expired and need an update.
func (w *Woolworths) ProductUpdateQueueWorker(output chan<- ProductID, maxAge time.Duration) {
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

func (w *Woolworths) NewDepartmentIDWorker(output chan<- DepartmentID) {
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

// This worker emits product IDs that don't currently exist in the local DB.
func (w *Woolworths) NewProductWorker(output chan<- WoolworthsProductInfo) {
	// TODO
	// output <- WoolworthsProductInfo{ID: 165262, Info: ProductInfo{}, Updated: time.Now().Add(-2 * w.productMaxAge)}
	// output <- WoolworthsProductInfo{ID: 187314, Info: ProductInfo{}, Updated: time.Now().Add(-2 * w.productMaxAge)}
	// output <- WoolworthsProductInfo{ID: 524336, Info: ProductInfo{}, Updated: time.Now().Add(-2 * w.productMaxAge)}
	// TODO Fix the below, it's busted in novel and interesting ways.
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
		if len(departments) > 0 {
			// If we have bootstrapped we don't need to check for new products very often.
			time.Sleep(1 * time.Hour)
		} else {
			time.Sleep(1 * time.Second)
		}

	}
}
