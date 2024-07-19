package woolworths

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	utils "github.com/tjhowse/aus_grocery_price_database/internal/utils"
)

func TestUnmarshal(t *testing.T) {
	// Read in the contents of data/example_product_info.json
	// and unmarshal it into a ProductInfo struct

	f, err := os.Open("data/187314.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	// Read the contents of the file
	body, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	// Unmarshal the contents of the file
	productInfo, err := UnmarshalProductInfo(body)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := "Woolworths Broccolini Bunch  Each", productInfo.Name; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}

}

// This mocks enough of the Woolworths API to test various stuff
func WoolworthsHTTPServer() *httptest.Server {
	var err error
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var responseData []byte
		if strings.HasPrefix(r.URL.Path, "/api/v3/ui/schemaorg/product/") {
			var productID int
			if _, err := fmt.Sscanf(r.URL.Path, "/api/v3/ui/schemaorg/product/%d", &productID); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if responseData, err = utils.ReadEntireFile(fmt.Sprintf("data/%d.json", productID)); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else if strings.HasPrefix(r.URL.Path, "/shop/browse/fruit-veg") {
			if responseData, err = utils.ReadEntireFile("data/fruit-veg.html"); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else if strings.HasPrefix(r.URL.Path, "/apis/ui/browse/category") {
			var categoryRequest CategoryRequestBody
			body, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			err = json.Unmarshal(body, &categoryRequest)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			responseFilename := fmt.Sprintf("data/category_%s_%d.json", categoryRequest.CategoryID, categoryRequest.PageNumber)
			if responseData, err = utils.ReadEntireFile(responseFilename); err != nil {
				fmt.Printf("Invalid category filename \"%s\", returning error\n", responseFilename)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(responseData)
	}))
}

func TestGetProductListPage(t *testing.T) {
	server := WoolworthsHTTPServer()
	defer server.Close()

	w := Woolworths{}
	w.Init(server.URL, ":memory:", PRODUCT_INFO_MAX_AGE)

	prodIDs, count, err := w.GetProductListPage("1-E5BEE36E", 1)
	if err != nil {
		t.Fatal(err)
	}
	if want, got := ProductID(133211), prodIDs[0]; want != got {
		t.Errorf("Expected %d, got %d", want, got)
	}

	if want, got := 470, count; want != got {
		t.Errorf("Expected a total product count of %d, got %d", want, got)
	}
}

func TestGetProductInfo(t *testing.T) {
	server := WoolworthsHTTPServer()
	defer server.Close()

	w := Woolworths{}
	w.Init(server.URL, ":memory:", PRODUCT_INFO_MAX_AGE)

	tests := map[int]string{
		187314: "Woolworths Broccolini Bunch  Each",
		165262: "Driscoll's Raspberries Punnet 125g Punnet",
		524336: "Woolworths Baby Spinach Spinach 280g",
	}

	for id, want := range tests {

		productInfo, err := w.GetProductInfo(ProductID(id))
		if err != nil {
			t.Fatal(err)
		}

		if productInfo.Info.Name != want {
			t.Errorf("Expected %s, got %s", want, productInfo.Info.Name)
		}
	}
}

func TestExtractDepartmentIDs(t *testing.T) {
	f, err := os.Open("data/fruit-veg.html")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	// Read the contents of the file
	body, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	departmentIDs, err := ExtractDepartmentIDs(body)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := 16, len(departmentIDs); want != got {
		t.Errorf("Expected %d items, got %d", want, got)
	}

	if want, got := DepartmentID("1-E5BEE36E"), departmentIDs[0]; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
}

func TestGetDepartmentIDs(t *testing.T) {
	server := WoolworthsHTTPServer()
	defer server.Close()

	w := Woolworths{}
	w.Init(server.URL, ":memory:", PRODUCT_INFO_MAX_AGE)

	departmentIDs, err := w.GetDepartmentIDs()
	if err != nil {
		t.Fatal(err)
	}
	if want, got := DepartmentID("1-E5BEE36E"), departmentIDs[0]; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
	if want, got := 16, len(departmentIDs); want != got {
		t.Errorf("Expected %d departments, got %d", want, got)
	}
}

func TestExtractTotalRecordCount(t *testing.T) {
	body, err := utils.ReadEntireFile("data/category_1-E5BEE36E_1.json")
	if err != nil {
		t.Fatal(err)
	}

	totalRecordCount, err := ExtractTotalRecordCount(body)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := 470, totalRecordCount; want != got {
		t.Errorf("Expected %d, got %d", want, got)
	}
}

func TestGetProductsFromDepartment(t *testing.T) {
	server := WoolworthsHTTPServer()
	defer server.Close()

	w := Woolworths{}
	err := w.Init(server.URL, ":memory:", PRODUCT_INFO_MAX_AGE)
	if err != nil {
		t.Fatal(err)
	}

	productIDs, err := w.GetProductsFromDepartment("1-E5BEE36E")
	if err != nil {
		t.Fatal(err)
	}
	if want, got := 38, len(productIDs); want != got {
		t.Errorf("Expected %d items, got %d", want, got)
	}
}

func TestProductInfoFetchingWorker(t *testing.T) {
	server := WoolworthsHTTPServer()
	defer server.Close()

	w := Woolworths{}
	err := w.Init(server.URL, ":memory:", PRODUCT_INFO_MAX_AGE)
	if err != nil {
		t.Fatal(err)
	}

	productInfoChannel := make(chan WoolworthsProductInfo)
	productsThatNeedAnUpdateChannel := make(chan ProductID)
	go w.ProductInfoFetchingWorker(productsThatNeedAnUpdateChannel, productInfoChannel)

	productsThatNeedAnUpdateChannel <- 187314

	select {
	case productInfo := <-productInfoChannel:
		if want, got := "Woolworths Broccolini Bunch  Each", productInfo.Info.Name; want != got {
			t.Errorf("Expected %s, got %s", want, got)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timed out waiting for product info")
	}

	// Set up a pipe to use as a substitute for stdout for the logger
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	defer writer.Close()

	// Tell slog to log to the pipe instead of stdout
	slog.SetDefault(slog.New(slog.NewTextHandler(writer, &slog.HandlerOptions{Level: slog.LevelDebug})))

	// Give it a bogus product that doesn't exist in the mocked webserver.
	productsThatNeedAnUpdateChannel <- 999999

	// Ensure we don't get a productInfo from the worker.
	select {
	case productInfo := <-productInfoChannel:
		t.Fatalf("Expected nothing, got %v", productInfo)
	case <-time.After(100 * time.Millisecond):
	}

	// Set up a semaphore for the goroutine to signal to us that it found the log message
	foundLogMessage := make(chan struct{})

	// Run a goroutine to scan the output of the logger for the expected message
	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "failed to get product info") {
				foundLogMessage <- struct{}{}
			}
		}
	}()

	// Wait for the log message to be found, or timeout.
	select {
	case <-foundLogMessage:
	case <-time.After(1000 * time.Millisecond):
		t.Fatal("Timed out waiting for log message")
	}

	// If this test seems tortuous, I agree. I'd like an easier way to capture log output from tests.
}
