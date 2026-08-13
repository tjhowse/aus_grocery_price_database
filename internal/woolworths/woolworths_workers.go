package woolworths

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

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
			w.logger.Error(fmt.Sprintf("Error getting department IDs from web: %v", err))
		}

		// Read the department list from the DB.
		departmentInfosFromDB, err := w.loadDepartmentInfoList()
		if err != nil {
			w.logger.Error(fmt.Sprintf("Error loading department IDs from DB: %v", err))
		}

		// Compare the two lists and output any new department IDs.
		for _, webDepartmentID := range departmentsFromWeb {
			if dept := departmentInSlice(webDepartmentID, departmentInfosFromDB); dept == nil {
				w.logger.Info("New department ID", "ID", webDepartmentID.NodeID, "Description", webDepartmentID.Description)
				output <- webDepartmentID
			} else {
				if dept.ProductCount != webDepartmentID.ProductCount {
					w.logger.Info("Department flagged for update", "oldProductCount", dept.ProductCount, "newProductCount", webDepartmentID.ProductCount)
					output <- webDepartmentID
				}
			}
		}
		// We don't need to check for departments very often.
		time.Sleep(1 * time.Hour)
	}
}

// productListPageWorker reads departmentPage structs from the input channel, fetches the product list page from the web,
// and writes the updated product data to the DB, transactionfully.
func (w *Woolworths) productListPageWorker(input <-chan departmentPage) {
	for dp := range input {
		w.logger.Debug("Getting product list page", "departmentID", dp.ID, "page", dp.page)
		products, err := w.getProductInfoFromListPage(dp)
		if err != nil {
			w.logger.Error(fmt.Sprintf("Error getting product info extended: %v", err))
			continue
		}
		tx, err := w.db.Begin()
		if err != nil {
			w.logger.Error(fmt.Sprintf("Error starting transaction: %v", err))
			continue
		}
		var skippedProductCount int
		for _, product := range products {
			// Skip products with zero price. Assume something went wrong.
			if product.Info.Price.Equal(decimal.Zero) {
				skippedProductCount++
				continue
			}
			product.departmentID = dp.ID
			err := w.saveProductInfo(tx, product)
			if err != nil {
				w.logger.Error(fmt.Sprintf("Error inserting product info: %v", err))
				continue
			}
		}
		err = tx.Commit()
		if err != nil {
			w.logger.Error(fmt.Sprintf("Error committing transaction: %v", err))
		}
		if skippedProductCount > 0 {
			w.logger.Debug("Skipped products with zero price", "skippedProductCount", skippedProductCount)
		}
	}
}

// departmentPageUpdateQueueWorker generates a stream of departmentPage structs that are due for an update
func (w *Woolworths) departmentPageUpdateQueueWorker(output chan<- departmentPage, maxAge time.Duration) {
	for {
		departmentInfos, err := w.loadDepartmentInfoList()
		if err != nil {
			w.logger.Error("error loading department IDs. Trying again soon.", "error", err)
			time.Sleep(1 * time.Minute)
			continue
		}
		for _, departmentInfo := range departmentInfos {
			if time.Since(departmentInfo.Updated) < maxAge {
				w.logger.Debug("Skipping update of department", "ID", departmentInfo.NodeID, "UpdatedAgo", time.Since(departmentInfo.Updated))
				continue
			}
			w.logger.Debug("Checking department", "ID", departmentInfo.NodeID, "Updated", departmentInfo.Updated)

			productCount := 0
			for productCount < departmentInfo.ProductCount {
				productCount += PRODUCTS_PER_PAGE
				w.logger.Debug("Adding department page to queue", "ID", departmentInfo.NodeID, "page", productCount/PRODUCTS_PER_PAGE)
				output <- departmentPage{
					ID:   departmentInfo.NodeID,
					page: productCount / PRODUCTS_PER_PAGE,
				}
			}
			// Save this department back to the DB to refresh its updated time.
			departmentInfo.Updated = time.Now()
			err := w.saveDepartment(departmentInfo)
			if err != nil {
				w.logger.Error("error saving department info", "error", err)
			}
			w.logger.Info("Updated department", "store", "Woolworths", "department", departmentInfo.Description)
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
			w.logger.Debug("New department", "ID", newDepartmentInfo.NodeID, "Description", newDepartmentInfo.Description, "ProductCount", newDepartmentInfo.ProductCount)
			// Update the departmentIDs table with the new department ID
			// Set the updated time in the past to force an update on the next poll.
			newDepartmentInfo.Updated = time.Now().Add(-2 * w.productMaxAge)
			err := w.saveDepartment(newDepartmentInfo)
			if err != nil {
				w.logger.Error(fmt.Sprintf("Error saving department ID: %v", err))
				continue
			}
			w.logger.Debug("Saved department", "ID", newDepartmentInfo.NodeID)

		case <-cancel:
			w.logger.Info("Exiting scheduler")
			return
		}
	}
}
