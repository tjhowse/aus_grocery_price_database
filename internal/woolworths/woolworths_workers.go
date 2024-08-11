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
		var actualUpdateInterval time.Duration
		updateTime := time.Now()

		transaction, err = w.db.Begin()
		if err != nil {
			slog.Error(fmt.Sprintf("Error starting transaction: %v", err))
			time.Sleep(10 * time.Second)
			continue
		}
		rows, err = transaction.Query(`	SELECT productID, updated FROM products
										WHERE updated < ?
										ORDER BY updated ASC
										LIMIT ?`, time.Now().Add(-maxAge), batchSize)
		if err != nil {
			if err != sql.ErrNoRows {
				slog.Error(fmt.Sprintf("Error getting productIDs due an update: %v", err))
			}
		} else {
			for rows.Next() {
				var productID productID
				var productUpdateTime time.Time
				err = rows.Scan(&productID, &productUpdateTime)
				if err != nil {
					slog.Error(fmt.Sprintf("Error scanning product ID: %v", err))
					continue
				}
				slog.Debug("Product ID needs an update", "productID", productID)
				productIDs = append(productIDs, productID)
				actualUpdateInterval += updateTime.Sub(productUpdateTime)
			}
		}
		if len(productIDs) > 0 {
			actualUpdateInterval /= time.Duration(len(productIDs))
			slog.Info("Actual update interval", "interval", actualUpdateInterval)
		}

		// Set the updated time for the selected products to now.
		// TODO see if this can be done with a productID IN (list) query.
		// I tried and failed, then used a transaction instead.
		for _, productID := range productIDs {
			_, err = transaction.Exec(`UPDATE products SET updated = ? WHERE productID = ?`, updateTime, productID)
			if err != nil {
				slog.Error(fmt.Sprintf("Error updating product info: %v", err))
				continue
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

func departmentInSlice(a departmentInfo, list []departmentInfo) *departmentInfo {
	for _, b := range list {
		if a.NodeID == b.NodeID {
			return &b
		}
	}
	return nil
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
			if dept := departmentInSlice(webDepartmentID, departmentInfosFromDB); dept == nil {
				slog.Info("New department ID", "ID", webDepartmentID.NodeID, "Description", webDepartmentID.Description)
				output <- webDepartmentID
			} else {
				if dept.ProductCount != webDepartmentID.ProductCount {
					slog.Info("Department flagged for update", "oldProductCount", dept.ProductCount, "newProductCount", webDepartmentID.ProductCount)
					output <- webDepartmentID
				}
			}
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

// productListPageWorker reads departmentPage structs from the input channel, fetches the product list page from the web,
// and writes the updated product data to the DB, transactionfully.
func (w *Woolworths) productListPageWorker(input <-chan departmentPage) {
	for dp := range input {
		slog.Debug("Getting product list page", "departmentID", dp.ID, "page", dp.page)
		products, err := w.getProductInfoExtendedFromListPage(dp)
		if err != nil {
			slog.Error(fmt.Sprintf("Error getting product info extended: %v", err))
			continue
		}
		transaction, err := w.db.Begin()
		if err != nil {
			slog.Error(fmt.Sprintf("Error starting transaction: %v", err))
			continue
		}
		for _, product := range products {
			product.departmentID = dp.ID
			err := w.saveProductInfoExtended(transaction, product)
			if err != nil {
				slog.Error(fmt.Sprintf("Error inserting product info: %v", err))
				continue
			}
		}
		err = transaction.Commit()
		if err != nil {
			slog.Error(fmt.Sprintf("Error committing transaction: %v", err))
		}
	}
}

// departmentPageUpdateQueueWorker generates a stream of departmentPage structs that are due for an update
func (w *Woolworths) departmentPageUpdateQueueWorker(output chan<- departmentPage, maxAge time.Duration) {
	for {
		departmentInfos, err := w.loadDepartmentInfoList()
		if err != nil {
			slog.Error("error loading department IDs. Trying again soon.", "error", err)
			time.Sleep(1 * time.Minute)
			continue
		}
		for _, departmentInfo := range departmentInfos {
			if time.Since(departmentInfo.Updated) < maxAge {
				slog.Debug("Skipping update of department", "ID", departmentInfo.NodeID, "UpdatedAgo", time.Since(departmentInfo.Updated))
				continue
			}
			slog.Debug("Checking department", "ID", departmentInfo.NodeID, "Updated", departmentInfo.Updated)

			productCount := 0
			for productCount < departmentInfo.ProductCount {
				productCount += PRODUCTS_PER_PAGE
				slog.Debug("Adding department page to queue", "ID", departmentInfo.NodeID, "page", productCount/PRODUCTS_PER_PAGE)
				output <- departmentPage{
					ID:   departmentInfo.NodeID,
					page: productCount / PRODUCTS_PER_PAGE,
				}
			}
			// Save this department back to the DB to refresh its updated time.
			departmentInfo.Updated = time.Now()
			err := w.saveDepartment(departmentInfo)
			if err != nil {
				slog.Error("error saving department info", "error", err)
			}
		}
		// We've done an update of all departments, so we don't need to check for new departments very often.
		time.Sleep(w.listingPageUpdateInterval)
	}
}

const DEFAULT_PRODUCT_UPDATE_BATCH_SIZE = 10

// Runs up all the workers and mediates data flowing between them.
// Currently all sqlite writes happen via this function. This may move
// off to a separate goroutine in the future.
func (w *Woolworths) Run(cancel chan struct{}) {
	// w.runIndividualPages(cancel)
	w.runExtended(cancel)
}

func (w *Woolworths) runIndividualPages(cancel chan struct{}) {

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
			// Set the updated time in the past to force an update on the next poll.
			newDepartmentInfo.Updated = time.Now().Add(-2 * w.productMaxAge)
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

func (w *Woolworths) runExtended(cancel chan struct{}) {
	departmentPageChannel := make(chan departmentPage)
	newDepartmentInfoChannel := make(chan departmentInfo)

	for i := 0; i < PRODUCT_INFO_WORKER_COUNT; i++ {
		go w.productListPageWorker(departmentPageChannel)
	}
	go w.newDepartmentInfoWorker(newDepartmentInfoChannel)
	go w.departmentPageUpdateQueueWorker(departmentPageChannel, w.productMaxAge)

	for {
		select {
		case newDepartmentInfo := <-newDepartmentInfoChannel:
			slog.Debug("New department", "ID", newDepartmentInfo.NodeID, "Description", newDepartmentInfo.Description, "ProductCount", newDepartmentInfo.ProductCount)
			// Update the departmentIDs table with the new department ID
			// Set the updated time in the past to force an update on the next poll.
			newDepartmentInfo.Updated = time.Now().Add(-2 * w.productMaxAge)
			err := w.saveDepartment(newDepartmentInfo)
			if err != nil {
				slog.Error(fmt.Sprintf("Error saving department ID: %v", err))
				continue
			}
			slog.Debug("Saved department", "ID", newDepartmentInfo.NodeID)

		case <-cancel:
			slog.Info("Exiting scheduler")
			return
		}
	}
}
