package woolworths

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

// This worker reads productIDs from the input channel, fetches the product info from the web,
// and writes the product info to the output channel.
func (w *Woolworths) productInfoFetchingWorker(input chan productID, output chan woolworthsProductInfo) {
	for id := range input {
		slog.Debug("Getting product", "id", id)
		info, err := w.getProductInfo(id)
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
func (w *Woolworths) productUpdateQueueWorker(output chan<- productID, maxAge time.Duration, batchSize int) {
	var err error
	for {
		var productIDs []productID
		var rows *sql.Rows
		var transaction *sql.Tx
		transaction, err = w.db.Begin()
		if err != nil {
			slog.Error(fmt.Sprintf("Error starting transaction: %v", err))
		}
		rows, err = transaction.Query(`	SELECT productID FROM products
									WHERE updated < ?
									ORDER BY updated ASC
									LIMIT ?`, time.Now().Add(-maxAge), batchSize)
		if err != nil {
			if err != sql.ErrNoRows {
				slog.Error(fmt.Sprintf("Error getting product ID: %v", err))
			}
		} else {
			for rows.Next() {
				var productID productID
				err = rows.Scan(&productID)
				if err != nil {
					slog.Error(fmt.Sprintf("Error scanning product ID: %v", err))
				}
				slog.Debug("Product ID needs an update", "productID", productID)
				productIDs = append(productIDs, productID)
			}
		}

		// Set the updated time for the selected products to now.
		// TODO see if this can be done with a productID IN (list) query.
		// I tried and failed, then used a transaction instead.
		for _, productID := range productIDs {
			_, err = transaction.Exec(`UPDATE products SET updated = ? WHERE productID = ?`, time.Now(), productID)
			if err != nil {
				slog.Error(fmt.Sprintf("Error updating product info: %v", err))
			}
		}

		err = transaction.Commit()
		if err != nil {
			slog.Error(fmt.Sprintf("Error committing transaction: %v", err))
		}

		for _, productID := range productIDs {
			output <- productID
		}
	}
}

func (w *Woolworths) filterProductIDs(productIDs []productID) []productID {
	var filtered []productID

	productSet := map[productID]bool{"133211": true, "134034": true, "105919": true, "144607": true, "208895": true, "135306": true, "144329": true, "134681": true, "170225": true, "169438": true, "135344": true, "120080": true, "135369": true, "829107": true, "144497": true, "130935": true, "149864": true, "149620": true, "147071": true, "137102": true, "137130": true, "157649": true, "120384": true, "259450": true, "155003": true, "314075": true, "713429": true, "727144": true, "147603": true, "144336": true, "829360": true, "165262": true, "310968": true, "154340": true, "187314": true, "262783": true}

	for _, productID := range productIDs {
		if _, ok := productSet[productID]; ok {
			filtered = append(filtered, productID)
		}
	}
	return filtered
}

// This worker emits a stream of new department IDs that don't currently exist in the database.
func (w *Woolworths) newDepartmentInfoWorker(output chan<- departmentInfo) {
	for {
		// Read the department list from the web...
		departmentsFromWeb, err := w.getDepartmentInfos()
		if err != nil {
			slog.Error(fmt.Sprintf("Error getting department IDs from web: %v", err))
		}

		// Read the department list from the DB.
		departmentInfosFromDB, err := w.loadDepartmentInfoList()
		if err != nil {
			slog.Error(fmt.Sprintf("Error loading department IDs from DB: %v", err))
		}

		// Compare the two lists and output any new department IDs.
		for _, webDepartmentID := range departmentsFromWeb {
			found := false
			for _, departmentInfoFromDB := range departmentInfosFromDB {
				if webDepartmentID.NodeID == departmentInfoFromDB.NodeID {
					found = true
					break
				}
			}
			if found {
				continue
			}
			if w.filterDepartments {
				if !w.filteredDepartmentIDsSet[webDepartmentID.NodeID] {
					continue
				}
			}

			output <- webDepartmentID
		}
		// We don't need to check for departments very often.
		time.Sleep(1 * time.Hour)
	}
}

// This worker emits product IDs that don't currently exist in the local DB.
func (w *Woolworths) newProductWorker(output chan<- woolworthsProductInfo) {
	for {
		departmentInfos, err := w.loadDepartmentInfoList()
		if err != nil {
			slog.Error("error loading department IDs. Trying again soon.", "error", err)
			// Try again in ten minutes.
			time.Sleep(1 * time.Minute)
			continue
		}
		for _, departmentInfo := range departmentInfos {
			products, err := w.getProductsFromDepartment(departmentInfo.NodeID)
			if err != nil {
				slog.Error("error getting products from department. Trying again later.", "error", err)
				time.Sleep(5 * time.Second)
				continue
			}

			if w.filterProducts {
				products = w.filterProductIDs(products)
			}

			for _, productID := range products {
				alreadyExists, err := w.checkIfKnownProductID(productID)
				if err != nil {
					slog.Error("error checking if known product ID", "error", err)
					continue
				}
				if alreadyExists {
					continue
				}
				output <- woolworthsProductInfo{
					ID:                    productID,
					departmentID:          departmentInfo.NodeID,
					departmentDescription: departmentInfo.Description,
					Info:                  productInfo{},
					Updated:               time.Now().Add(-2 * w.productMaxAge)}
			}
		}
		if len(departmentInfos) > 0 {
			// If we have bootstrapped we don't need to check for new departments very often.
			time.Sleep(10 * time.Minute)
		} else {
			time.Sleep(1 * time.Second)
		}

	}
}

const DEFAULT_PRODUCT_UPDATE_BATCH_SIZE = 10

// Runs up all the workers and mediates data flowing between them.
// Currently all sqlite writes happen via this function. This may move
// off to a separate goroutine in the future.
func (w *Woolworths) Run(cancel chan struct{}) {

	productInfoChannel := make(chan woolworthsProductInfo)
	productsThatNeedAnUpdateChannel := make(chan productID)
	newDepartmentInfoChannel := make(chan departmentInfo)
	for i := 0; i < PRODUCT_INFO_WORKER_COUNT; i++ {
		go w.productInfoFetchingWorker(productsThatNeedAnUpdateChannel, productInfoChannel)
	}
	go w.productUpdateQueueWorker(productsThatNeedAnUpdateChannel, w.productMaxAge, DEFAULT_PRODUCT_UPDATE_BATCH_SIZE)
	go w.newProductWorker(productInfoChannel)
	go w.newDepartmentInfoWorker(newDepartmentInfoChannel)

	for {
		var err error
		select {
		case productInfoUpdate := <-productInfoChannel:
			slog.Debug("Read from productInfoChannel", "name", productInfoUpdate.Info.Name)
			// Update the product info in the DB
			err = w.saveProductInfo(productInfoUpdate)
			if err != nil {
				slog.Error(fmt.Sprintf("Error saving product info: %v", err))
			}
		case newDepartmentInfo := <-newDepartmentInfoChannel:
			slog.Debug("New department", "ID", newDepartmentInfo.NodeID, "Description", newDepartmentInfo.Description)
			// Update the departmentIDs table with the new department ID
			err := w.saveDepartment(newDepartmentInfo)
			if err != nil {
				slog.Error(fmt.Sprintf("Error saving department ID: %v", err))
			}
		case <-cancel:
			slog.Info("Exiting scheduler")
			return
		}
	}

}
