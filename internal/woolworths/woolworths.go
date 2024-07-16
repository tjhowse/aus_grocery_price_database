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

func ExtractStockCodes(body []byte) ([]int, error) {
	stockCodeRegex := regexp.MustCompile(`"Stockcode":(\d*),`)
	stockCodeMatches := stockCodeRegex.FindAllStringSubmatch(string(body), -1)
	if len(stockCodeMatches) == 0 {
		return []int{}, fmt.Errorf("no stock codes found")
	}

	stockCodes := []int{}
	for _, code := range stockCodeMatches {
		id, err := strconv.Atoi(code[1])
		if err != nil {
			return []int{}, err
		} else {
			stockCodes = append(stockCodes, id)
		}
	}

	return stockCodes, nil
}

// This extracts a substring out of the fruit-veg page and uses a regex to find
// the list of department IDs within. It decodes this list as json.
func ExtractDepartmentIDs(body []byte) ([]DepartmentID, error) {
	departmentIDListRegex := regexp.MustCompile(`{"Group":"lists","Name":"includedDepartmentIds","Value":\[.*?\]}`)
	departmentIDListMatches := departmentIDListRegex.FindAllStringSubmatch(string(body), -1)
	if len(departmentIDListMatches) == 0 {
		return []DepartmentID{}, fmt.Errorf("no department IDs found")
	}

	var department Department
	err := json.Unmarshal([]byte(departmentIDListMatches[0][0]), &department)
	if err != nil {
		return []DepartmentID{}, err
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

func (w *Woolworths) GetProductList() ([]ProductID, error) {
	var url string

	prodIDs := []ProductID{}

	requestBody := `{"categoryId":"1-E5BEE36E","pageNumber":1,"pageSize":36,"sortType":"TraderRelevance","url":"/shop/browse/fruit-veg","location":"/shop/browse/fruit-veg","formatObject":"{\"name\":\"Fruit & Veg\"}","isSpecial":false,"isBundle":false,"isMobile":false,"filters":[],"token":"","gpBoost":0,"isHideUnavailableProducts":false,"isRegisteredRewardCardPromotion":false,"enableAdReRanking":false,"groupEdmVariants":true,"categoryVersion":"v2"}`

	url = fmt.Sprintf("%s/apis/ui/browse/category", w.baseURL)
	if req, err := http.NewRequest("POST", url, bytes.NewBufferString(requestBody)); err != nil {
		return prodIDs, err
	} else {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:127.0) Gecko/20100101 Firefox/127.0")
		req.Header.Set("Accept", "application/json, text/plain, */*")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Request-Id", "|b14af797522740e5a25290ac283f739d.037da5c5e87f4706")
		resp, err := w.client.Do(req)
		if err != nil {
			return prodIDs, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return prodIDs, err
		}

		stockCodes, err := ExtractStockCodes(body)
		if err != nil {
			return prodIDs, err
		}

		for _, code := range stockCodes {
			prodIDs = append(prodIDs, ProductID(code))
		}
	}

	return prodIDs, nil
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
		return ProductInfo{}, err
	}

	return productInfo, nil
}
