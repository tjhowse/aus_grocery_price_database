package woolworths

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

func (w *Woolworths) ProductInfoFetchingWorker(input chan ProductID, output chan WoolworthsProductInfo) {
	for id := range input {
		slog.Debug("Getting product", "id", id)
		info, err := w.GetProductInfo(id)
		if err != nil {
			// Log an error to update this info, but still flag it as updated.
			// This prevents dud product IDs clogging up the system, at the cost
			// of potentially missing an update occasionally.
			slog.Error(fmt.Sprintf("Error getting product info: %v", err))
		}
		info.Updated = time.Now()
		output <- info
	}
}

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
			slog.Debug("Feeding a product ID out of ProductUpdateQueueWorker", "productID", productID)
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
			// If we have bootstrapped we don't need to check for new departments very often.
			time.Sleep(1 * time.Hour)
		} else {
			time.Sleep(1 * time.Second)
		}

	}
}

// Runs up all the workers and mediates data flowing between them.
// Currently all sqlite writes happen via this function. This may move
// off to a separate goroutine in the future.
func (w *Woolworths) RunScheduler(cancel chan struct{}) {

	productInfoChannel := make(chan WoolworthsProductInfo)
	productsThatNeedAnUpdateChannel := make(chan ProductID)
	newDepartmentIDsChannel := make(chan DepartmentID)
	for i := 0; i < PRODUCT_INFO_WORKER_COUNT; i++ {
		go w.ProductInfoFetchingWorker(productsThatNeedAnUpdateChannel, productInfoChannel)
	}
	go w.ProductUpdateQueueWorker(productsThatNeedAnUpdateChannel, w.productMaxAge)
	go w.NewProductWorker(productInfoChannel)
	go w.NewDepartmentIDWorker(newDepartmentIDsChannel)

	for {
		slog.Debug("Heartbeat")
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
			err := w.SaveDepartment(newDepartmentID)
			if err != nil {
				slog.Error(fmt.Sprintf("Error saving department ID: %v", err))
			}
		case <-cancel:
			slog.Info("Exiting scheduler")
			return
		}
	}

}
