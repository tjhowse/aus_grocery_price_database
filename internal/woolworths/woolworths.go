package woolworths

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"os"
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

func (w *Woolworths) GetProductList() ([]ProductID, error) {
	var url string

	prodIDs := []ProductID{}

	url = fmt.Sprintf("%s/shop/browse/fruit-veg", w.baseURL)
	if req, err := http.NewRequest("GET", url, nil); err != nil {
		return prodIDs, err
	} else {
		w.client.Do(req)
	}

	requestBody := `{"categoryId":"1-E5BEE36E","pageNumber":1,"pageSize":36,"sortType":"TraderRelevance","url":"/shop/browse/fruit-veg","location":"/shop/browse/fruit-veg","formatObject":"{\"name\":\"Fruit & Veg\"}","isSpecial":false,"isBundle":false,"isMobile":false,"filters":[],"token":"","gpBoost":0,"isHideUnavailableProducts":false,"isRegisteredRewardCardPromotion":false,"enableAdReRanking":false,"groupEdmVariants":true,"categoryVersion":"v2"}`

	url = fmt.Sprintf("%s/apis/ui/browse/category", w.baseURL)
	if req, err := http.NewRequest("POST", url, bytes.NewBufferString(requestBody)); err != nil {
		return prodIDs, err
	} else {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:127.0) Gecko/20100101 Firefox/127.0")
		req.Header.Set("Accept", "application/json, text/plain, */*")
		req.Header.Set("Content-Type", "application/json")
		// req.Header.Set("Request-Id", "|f14af797522740e5a25290ac283f739d.037da5c5e87f4706")
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

		// write the response to a file for manual inspection.

		f, err := os.Create("category.json")

		if err != nil {
			return prodIDs, err
		}
		defer f.Close()
		f.Write(body)

		// Now we have a page of products. If I wanted I could parse it all into a big struct,
		// but I only really care about getting the Stockcode out of it. So I'll just do that.

		stockCodeRegex := regexp.MustCompile(`"Stockcode":(\d*),`)
		stockCodeMatches := stockCodeRegex.FindAllStringSubmatch(string(body), -1)
		if len(stockCodeMatches) == 0 {
			return prodIDs, fmt.Errorf("no stock codes found")
		}

		for _, code := range stockCodeMatches {
			id, err := strconv.Atoi(code[1])
			if err != nil {
				return prodIDs, err
			} else {
				prodIDs = append(prodIDs, ProductID{ID: id})
			}
		}
	}

	return prodIDs, nil
}

// This queries the Woolworths API to get the product information
// using the WOOLWORTHS_PRODUCT_URL_PREFIX prefix.
func (w *Woolworths) GetProductInfo(id ProductID) (ProductInfo, error) {
	slog.Debug(fmt.Sprintf("Base URL: %s", w.baseURL))
	url := fmt.Sprintf(WOOLWORTHS_PRODUCT_URL_FORMAT, w.baseURL, id.ID)

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
