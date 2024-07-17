package woolworths

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strconv"
	"time"

	"golang.org/x/time/rate"
)

const WOOLWORTHS_PRODUCT_URL_FORMAT = "%s/api/v3/ui/schemaorg/product/%d"

type Woolworths struct {
	baseURL   string
	client    *RLHTTPClient
	cookieJar *cookiejar.Jar
}

func (w *Woolworths) Init(baseURL string) {
	var err error
	w.cookieJar, err = cookiejar.New(nil)
	if err != nil {
		slog.Error(fmt.Sprintf("Error creating cookie jar: %v", err))
	}
	w.baseURL = baseURL
	w.client = &RLHTTPClient{
		client: &http.Client{
			Jar: w.cookieJar,
		},
		Ratelimiter: rate.NewLimiter(rate.Every(1*time.Second), 1),
	}
}

func ExtractStockCodes(body CategoryData) ([]int, error) {
	stockCodeRegex := regexp.MustCompile(`"Stockcode":(\d*),`)
	stockCodeMatches := stockCodeRegex.FindAllStringSubmatch(string(body), -1)
	if len(stockCodeMatches) == 0 {
		return []int{}, fmt.Errorf("no stock codes found")
	}

	stockCodes := []int{}
	for _, code := range stockCodeMatches {
		id, err := strconv.Atoi(code[1])
		if err != nil {
			return []int{}, fmt.Errorf("failed to parse stock code: %w", err)
		} else {
			stockCodes = append(stockCodes, id)
		}
	}

	return stockCodes, nil
}

// This extracts a substring out of the fruit-veg page and uses a regex to find
// the list of department IDs within. It decodes this list as json.
func ExtractDepartmentIDs(body FruitVegPage) ([]DepartmentID, error) {
	departmentIDListRegex := regexp.MustCompile(`{"Group":"lists","Name":"includedDepartmentIds","Value":\[.*?\]}`)
	departmentIDListMatches := departmentIDListRegex.FindAllStringSubmatch(string(body), -1)
	if len(departmentIDListMatches) == 0 {
		return []DepartmentID{}, fmt.Errorf("no department IDs found")
	}

	var department Department
	err := json.Unmarshal([]byte(departmentIDListMatches[0][0]), &department)
	if err != nil {
		return []DepartmentID{}, fmt.Errorf("failed to unmarshal department information: %w", err)
	}

	return department.Value, nil
}

func (w *Woolworths) GetDepartmentIDs() ([]DepartmentID, error) {
	departmentIDs := []DepartmentID{}
	url := fmt.Sprintf("%s/shop/browse/fruit-veg", w.baseURL)
	if req, err := http.NewRequest("GET", url, nil); err != nil {
		return departmentIDs, err
	} else {
		resp, err := w.client.Do(req)
		if err != nil {
			return departmentIDs, err
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return departmentIDs, err
		}
		departmentIDs, err = ExtractDepartmentIDs(body)
		if err != nil {
			return departmentIDs, err
		}
		return departmentIDs, nil
	}
}

func ExtractTotalRecordCount(body CategoryData) (int, error) {
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

func BuildCategoryRequestBody(departmentID DepartmentID, pageNumber int) string {
	return fmt.Sprintf(`{"categoryId":"%s","pageNumber":%d,"pageSize":36,"sortType":"TraderRelevance","url":"/shop/browse/fruit-veg","location":"/shop/browse/fruit-veg","formatObject":"{\"name\":\"Fruit & Veg\"}","isSpecial":false,"isBundle":false,"isMobile":false,"filters":[],"token":"","gpBoost":0,"isHideUnavailableProducts":false,"isRegisteredRewardCardPromotion":false,"enableAdReRanking":false,"groupEdmVariants":true,"categoryVersion":"v2"}`, departmentID, pageNumber)
}

func (w *Woolworths) GetProductList() ([]ProductID, error) {

	prodIDs := []ProductID{}

	departmentIDs, err := w.GetDepartmentIDs()
	if err != nil {
		return prodIDs, err
	}

	// This is a long-running process. We probably don't want to split it into multiple
	// concurrent workers out of politeness to the Woolworths API. We only need to refresh
	// our product list once a day or so, so it's OK if it takes a while to run.
	for _, departmentID := range departmentIDs {
		ids, err := w.GetProductsFromDepartment(departmentID)
		if err != nil {
			return prodIDs, err
		}
		prodIDs = append(prodIDs, ids...)
	}

	return prodIDs, nil
}

func (w *Woolworths) GetProductsFromDepartment(department DepartmentID) ([]ProductID, error) {
	prodIDs := []ProductID{}
	page := 1

	for {
		ids, count, err := w.GetProductListPage(department, page)
		if err != nil {
			return prodIDs, err
		}
		prodIDs = append(prodIDs, ids...)
		if len(prodIDs) >= count {
			break
		}
		page++
	}

	return prodIDs, nil
}

func (w *Woolworths) GetProductListPage(department DepartmentID, page int) ([]ProductID, int, error) {
	var url string
	var totalCount int

	prodIDs := []ProductID{}

	requestBody := BuildCategoryRequestBody(department, page)

	url = fmt.Sprintf("%s/apis/ui/browse/category", w.baseURL)
	if req, err := http.NewRequest("POST", url, bytes.NewBufferString(requestBody)); err != nil {
		return prodIDs, 0, err
	} else {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:127.0) Gecko/20100101 Firefox/127.0")
		req.Header.Set("Accept", "application/json, text/plain, */*")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Request-Id", "|b14af797522740e5a25290ac283f739d.037da5c5e87f4706")
		resp, err := w.client.Do(req)
		if err != nil {
			return prodIDs, 0, fmt.Errorf("failed to get category data: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return prodIDs, 0, err
		}

		totalCount, err = ExtractTotalRecordCount(body)
		if err != nil {
			return prodIDs, 0, err
		}

		stockCodes, err := ExtractStockCodes(body)
		if err != nil {
			return prodIDs, 0, err
		}

		for _, code := range stockCodes {
			prodIDs = append(prodIDs, ProductID(code))
		}
	}

	return prodIDs, totalCount, nil
}

// This queries the Woolworths API to get the product information
// using the WOOLWORTHS_PRODUCT_URL_PREFIX prefix.
func (w *Woolworths) GetProductInfo(id ProductID) (ProductInfo, error) {
	slog.Debug(fmt.Sprintf("Base URL: %s", w.baseURL))
	url := fmt.Sprintf(WOOLWORTHS_PRODUCT_URL_FORMAT, w.baseURL, id)

	// Create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ProductInfo{}, err
	}

	// Dispatch the request
	resp, err := w.client.Do(req)
	if err != nil {
		return ProductInfo{}, err
	}

	// Parse the response
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ProductInfo{}, err
	}

	return UnmarshalProductInfo(body)
}

func UnmarshalProductInfo(body []byte) (ProductInfo, error) {
	var productInfo ProductInfo

	if err := json.Unmarshal(body, &productInfo); err != nil {
		return ProductInfo{}, fmt.Errorf("failed to unmarshal product info: %w", err)
	}

	return productInfo, nil
}
