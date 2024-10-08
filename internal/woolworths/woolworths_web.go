package woolworths

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

func extractStockCodes(body categoryData) ([]string, error) {
	stockCodeRegex := regexp.MustCompile(`"Stockcode":(\d*),`)
	stockCodeMatches := stockCodeRegex.FindAllStringSubmatch(string(body), -1)
	if len(stockCodeMatches) == 0 {
		return []string{}, fmt.Errorf("no stock codes found")
	}

	stockCodes := []string{}
	for _, code := range stockCodeMatches {
		stockCodes = append(stockCodes, code[1])
	}

	return stockCodes, nil
}

// This extracts a substring out of the fruit-veg page and uses a regex to find
// the list of department information within. It decodes this list as json.
func extractDepartmentInfos(body fruitVegPage) ([]departmentInfo, error) {
	// departmentInfoListRegex := regexp.MustCompile(`{"Group":"lists","Name":"includedDepartmentIds","Value":\[.*?\]}`)
	departmentInfoListRegex := regexp.MustCompile(`{"Categories":\[{"NodeId":"specialsgroup","Description":"Specials".*?]}`)
	departmentIDListMatches := departmentInfoListRegex.FindAllStringSubmatch(string(body), -1)
	if len(departmentIDListMatches) == 0 {
		return []departmentInfo{}, fmt.Errorf("no department IDs found")
	}

	var departmentList DepartmentCategoriesList
	err := json.Unmarshal([]byte(departmentIDListMatches[0][0]), &departmentList)
	if err != nil {
		return []departmentInfo{}, fmt.Errorf("failed to unmarshal department information: %w", err)
	}

	return departmentList.Categories, nil
}

func (w *Woolworths) getDepartmentInfos() ([]departmentInfo, error) {
	var req *http.Request
	var resp *http.Response
	var err error
	departmentInfos := []departmentInfo{}

	url := fmt.Sprintf("%s/shop/browse/fruit-veg", w.baseURL)
	if req, err = http.NewRequest("GET", url, nil); err != nil {
		return departmentInfos, err
	}
	resp, err = w.client.Do(req)
	if err != nil {
		return departmentInfos, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return departmentInfos, fmt.Errorf("failed to get category data: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return departmentInfos, err
	}
	departmentInfos, err = extractDepartmentInfos(body)
	if err != nil {
		return departmentInfos, err
	}
	departmentInfos = w.filterOutDepartments(departmentInfos)
	// Now we have to populate the product count, since the fruit-veg page doesn't have it.
	for i, departmentInfo := range departmentInfos {
		_, count, err := w.getProductIDsAndCountFromListPage(departmentInfo.NodeID, 1)
		if err != nil {
			slog.Warn("Failed to get product count for department", "department", departmentInfo.NodeID, "error", err)
			continue
		}
		departmentInfos[i].ProductCount = count
	}
	return departmentInfos, nil
}

func extractTotalRecordCount(body categoryData) (int, error) {
	totalRecordCountRegex := regexp.MustCompile(`"TotalRecordCount":(\d*),`)
	totalRecordCountMatches := totalRecordCountRegex.FindAllStringSubmatch(string(body), -1)
	if len(totalRecordCountMatches) == 0 {
		return 0, fmt.Errorf("total record count not found")
	}
	count, err := strconv.Atoi(totalRecordCountMatches[0][1])
	if err != nil {
		return 0, fmt.Errorf("failed to parse total record count: %w", err)
	}

	return count, nil
}

func buildCategoryRequestBody(departmentID departmentID, pageNumber int) (string, error) {
	// Example:
	// {
	// 	"categoryId": "adadsf",
	// 	"pageNumber": 1,
	// 	"pageSize": 36,
	// 	"sortType": "TraderRelevance",
	// 	"url": "/shop/browse/fruit-veg",
	// 	"location": "/shop/browse/fruit-veg",
	// 	"formatObject": "{\"name\":\"Fruit & Veg\"}",
	// 	"isSpecial": false,
	// 	"isBundle": false,
	// 	"isMobile": false,
	// 	"filters": [],
	// 	"token": "",
	// 	"gpBoost": 0,
	// 	"isHideUnavailableProducts": false,
	// 	"isRegisteredRewardCardPromotion": false,
	// 	"enableAdReRanking": false,
	// 	"groupEdmVariants": true,
	// 	"categoryVersion": "v2"
	// }
	pageData := categoryRequestBody{
		CategoryID:                      departmentID,
		PageNumber:                      pageNumber,
		PageSize:                        36,
		SortType:                        "TraderRelevance",
		URL:                             "/shop/browse/fruit-veg",
		Location:                        "/shop/browse/fruit-veg",
		FormatObject:                    "{\"name\":\"Fruit & Veg\"}",
		IsSpecial:                       false,
		IsBundle:                        false,
		IsMobile:                        false,
		Filters:                         []string{},
		Token:                           "",
		GPBoost:                         0,
		IsHideUnavailableProducts:       false,
		IsRegisteredRewardCardPromotion: false,
		EnableAdReRanking:               false,
		GroupEdmVariants:                true,
		CategoryVersion:                 "v2",
	}
	request, err := json.Marshal(pageData)
	if err != nil {
		return "", fmt.Errorf("error marshalling page data: %w", err)
	}
	return string(request), nil
}

// extractProductInfoFromProductListPage extracts the product info from the product list page
func extractProductInfoFromProductListPage(body []byte) ([]woolworthsProductInfo, error) {
	productInfos := []woolworthsProductInfo{}

	// Unmarshal body into a productListPage
	var productListPage productListPage
	if err := json.Unmarshal(body, &productListPage); err != nil {
		return productInfos, fmt.Errorf("failed to unmarshal product list page: %w", err)
	}

	// // Extract the product info from the productListPage
	for _, products := range productListPage.Bundles {
		if len(products.Products) != 1 {
			return productInfos, fmt.Errorf("expected 1 product in bundle, got %d", len(products.Products))
		}
		product := products.Products[0]
		// Not entirely happy about this, but I don't see a better way of getting the raw JSON
		// of just the product into the info struct.
		encoded, err := json.Marshal(product)
		if err != nil {
			slog.Warn("Error encoding product back to JSON for storage", "error", err)
			encoded = []byte("error re-encoding product")
		}
		productInfos = append(productInfos, woolworthsProductInfo{
			ID:                    productID(strconv.Itoa(product.Stockcode)),
			departmentID:          departmentID(product.AdditionalAttributes.PiesProductDepartmentNodeID),
			departmentDescription: product.AdditionalAttributes.Sapdepartmentname,
			RawJSON:               encoded,
			Info:                  product,
			Updated:               time.Now(),
		})
	}

	return productInfos, nil
}

// getProductListPage returns the bytes of the product list page for the given department and page number.
func (w *Woolworths) getProductListPage(department departmentID, page int) ([]byte, error) {

	var url string

	requestBody, err := buildCategoryRequestBody(department, page)
	if err != nil {
		return nil, err
	}
	slog.Debug("Requesting product info page", "department", department, "page", page)

	url = fmt.Sprintf("%s/apis/ui/browse/category", w.baseURL)
	if req, err := http.NewRequest("POST", url, bytes.NewBufferString(requestBody)); err != nil {
		return nil, err
	} else {
		// This is the minimal set of headers the request expects to see.
		// Note that the cookie jar must be full from previous requests
		// to the /shop/browse/* endpoint for this to work.
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:127.0) Gecko/20100101 Firefox/127.0")
		req.Header.Set("Accept", "application/json, text/plain, */*")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Request-Id", "|b14af797522740e5a25290ac283f739d.037da5c5e87f4706")
		resp, err := w.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to get category data: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to get data from category %s page %d: %s", department, page, resp.Status)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return body, nil
	}
}

// This queries the Woolworths API to get the product list for a department. It reads
// the specified page of that department's product list, returning the list of product
// IDs and the total number of products in the department.
func (w *Woolworths) getProductIDsAndCountFromListPage(department departmentID, page int) ([]productID, int, error) {
	var totalCount int

	prodIDs := []productID{}
	body, err := w.getProductListPage(department, page)
	if err != nil {
		return prodIDs, 0, err
	}

	totalCount, err = extractTotalRecordCount(body)
	if err != nil {
		return prodIDs, 0, err
	}

	stockCodes, err := extractStockCodes(body)
	if err != nil {
		return prodIDs, 0, err
	}

	for _, code := range stockCodes {
		prodIDs = append(prodIDs, productID(code))
	}
	return prodIDs, totalCount, nil
}

// getProductInfoFromListPage returns the product information from the department list page
func (w *Woolworths) getProductInfoFromListPage(dp departmentPage) ([]woolworthsProductInfo, error) {
	productInfos := []woolworthsProductInfo{}
	var body []byte
	var err error

	body, err = w.getProductListPage(dp.ID, dp.page)
	if err != nil {
		return productInfos, err
	}

	return extractProductInfoFromProductListPage(body)
}

// isDepartmentFilteredOut returns true if the department is in the filteredDepartmentIDsSet
func (w *Woolworths) isDepartmentFilteredOut(department departmentID) bool {
	if !w.filterDepartments {
		return false
	}
	_, ok := w.filteredDepartmentIDsSet[department]
	return !ok
}

// filterOutDepartments filters out the departments that are not in the filteredDepartmentIDsSet
func (w *Woolworths) filterOutDepartments(departments []departmentInfo) []departmentInfo {
	filteredDepartments := []departmentInfo{}
	for _, dp := range departments {
		if !w.isDepartmentFilteredOut(dp.NodeID) {
			filteredDepartments = append(filteredDepartments, dp)
		}
	}
	return filteredDepartments
}
