package coles

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
)

const PRODUCTS_PER_PAGE = 48

func departmentInSlice(a departmentInfo, list []departmentInfo) *departmentInfo {
	for _, b := range list {
		if a.SeoToken == b.SeoToken {
			return &b
		}
	}
	return nil
}

// newDepartmentInfoWorker is a worker that monitors for new departments and writes them to the DB.
func (c *Coles) newDepartmentInfoWorker() {
	for {

		// Read the department list from the web...
		departmentsFromWeb, err := c.getDepartmentInfos()
		if err != nil {
			slog.Error(fmt.Sprintf("Error getting department IDs from web: %v", err))
		}

		// Read the department list from the DB.
		departmentInfosFromDB, err := c.loadDepartmentInfoList()
		if err != nil {
			slog.Error(fmt.Sprintf("Error loading department IDs from DB: %v", err))
		}

		// Compare the two lists and output any new department IDs.
		for _, webDepartmentInfo := range departmentsFromWeb {
			update := false
			if dept := departmentInSlice(webDepartmentInfo, departmentInfosFromDB); dept == nil {
				slog.Info("New department ID", "ID", webDepartmentInfo.SeoToken, "Description", webDepartmentInfo.Name)
				update = true
			} else {
				if dept.ProductCount != webDepartmentInfo.ProductCount {
					slog.Info("Department flagged for update", "oldProductCount", dept.ProductCount, "newProductCount", webDepartmentInfo.ProductCount)
					update = true
				}
			}
			if update {
				// Save the department to the DB.
				// Set the update time to the past so we force an update on the next poll.
				webDepartmentInfo.Updated = time.Now().Add(-2 * c.productMaxAge)
				err := c.saveDepartment(webDepartmentInfo)
				if err != nil {
					slog.Error(fmt.Sprintf("Error saving department ID to DB: %v", err))
				}
			}
		}

		// We don't need to check for departments very often.
		time.Sleep(1 * time.Hour)

		// Update this every so often.
		if err := c.updateAPIVersion(); err != nil {
			slog.Error("error updating API version", "error", err)
		}

	}
}

// departmentPageUpdateQueueWorker generates a stream of departmentPage structs that are due for an update
func (c *Coles) departmentPageUpdateQueueWorker(output chan<- departmentPage, maxAge time.Duration) {
	for {
		departmentInfos, err := c.loadDepartmentInfoList()
		if err != nil {
			slog.Error("error loading department IDs. Trying again soon.", "error", err)
			time.Sleep(1 * time.Minute)
			continue
		}
		for _, departmentInfo := range departmentInfos {
			if c.filterDepartments {
				_, ok := c.filteredDepartmentIDsSet[departmentInfo.SeoToken]
				if !ok {
					slog.Debug("Skipping department", "SeoToken", departmentInfo.SeoToken)
					continue
				}
			}

			if time.Since(departmentInfo.Updated) < maxAge {
				slog.Debug("Skipping update of department", "SeoToken", departmentInfo.SeoToken, "UpdatedAgo", time.Since(departmentInfo.Updated))
				continue
			}
			slog.Debug("Checking department", "ID", departmentInfo.SeoToken, "Updated", departmentInfo.Updated)

			productCount := 0
			for productCount < departmentInfo.ProductCount {
				productCount += PRODUCTS_PER_PAGE
				slog.Debug("Adding department page to queue", "SeoToken", departmentInfo.SeoToken, "page", productCount/PRODUCTS_PER_PAGE)
				output <- departmentPage{
					ID:   departmentInfo.SeoToken,
					page: productCount / PRODUCTS_PER_PAGE,
				}
			}
			// Save this department back to the DB to refresh its updated time.
			departmentInfo.Updated = time.Now()
			err := c.saveDepartment(departmentInfo)
			if err != nil {
				slog.Error("error saving department info", "error", err)
			}
		}
		// We've done an update of all departments, so we don't need to check for new departments very often.
		time.Sleep(c.listingPageUpdateInterval)
	}
}

// productListPageWorker reads departmentPage structs from the input channel, fetches the product list page from the web,
// and writes the updated product data to the DB, transactionfully.
func (w *Coles) productListPageWorker(input <-chan departmentPage) {
	for dp := range input {
		slog.Debug("Getting product list page", "departmentID", dp.ID, "page", dp.page)
		products, _, err := w.getProductsAndTotalCountForCategoryPage(dp)
		if err != nil {
			slog.Error(fmt.Sprintf("Error getting product info extended: %v", err))
			continue
		}
		tx, err := w.db.Begin()
		if err != nil {
			slog.Error(fmt.Sprintf("Error starting transaction: %v", err))
			continue
		}
		var skippedProductCount int
		for _, product := range products {
			// Skip products with zero price. Assume something went wrong.
			if product.Info.Pricing.Now.Equal(decimal.Zero) {
				skippedProductCount++
				continue
			}
			product.departmentID = dp.ID
			err := w.saveProductInfo(tx, product)
			if err != nil {
				slog.Error(fmt.Sprintf("Error inserting product info: %v", err))
				continue
			}
		}
		err = tx.Commit()
		if err != nil {
			slog.Error(fmt.Sprintf("Error committing transaction: %v", err))
		}
		if skippedProductCount > 0 {
			slog.Warn("Skipped products with zero price", "skippedProductCount", skippedProductCount)
		}
	}
}
