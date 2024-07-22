package woolworths

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	utils "github.com/tjhowse/aus_grocery_price_database/internal/utils"
)

var woolworthsServer = WoolworthsHTTPServer()

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

	filesToLoad := []string{
		"data/187314.json",
		"data/165262.json",
		"data/524336.json",
		"data/category_1-E5BEE36E_1.json",
		"data/category_1-E5BEE36E_2.json",
		"data/fruit-veg.html",
	}
	fileContents := make(map[string][]byte)
	for _, filename := range filesToLoad {
		fileContents[filename], err = utils.ReadEntireFile(filename)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to read file %s: %v\n", filename, err))
		}
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var responseFilename string
		if strings.HasPrefix(r.URL.Path, "/api/v3/ui/schemaorg/product/") {
			var productID int
			if _, err := fmt.Sscanf(r.URL.Path, "/api/v3/ui/schemaorg/product/%d", &productID); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			responseFilename = fmt.Sprintf("data/%d.json", productID)
		} else if strings.HasPrefix(r.URL.Path, "/shop/browse/fruit-veg") {
			responseFilename = "data/fruit-veg.html"
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
			responseFilename = fmt.Sprintf("data/category_%s_%d.json", categoryRequest.CategoryID, categoryRequest.PageNumber)
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if responseData, knownFile := fileContents[responseFilename]; !knownFile {
			slog.Error("Simulated woolworths server can't find requested file.", "filename", responseFilename)
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(responseData)
		}
	}))
}

func TestGetProductListPage(t *testing.T) {
	w := Woolworths{}
	w.Init(woolworthsServer.URL, ":memory:", PRODUCT_INFO_MAX_AGE)

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
	w := Woolworths{}
	w.Init(woolworthsServer.URL, ":memory:", PRODUCT_INFO_MAX_AGE)

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
	w := Woolworths{}
	w.Init(woolworthsServer.URL, ":memory:", PRODUCT_INFO_MAX_AGE)

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

	w := Woolworths{}
	err := w.Init(woolworthsServer.URL, ":memory:", PRODUCT_INFO_MAX_AGE)
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

func TestGetProductList(t *testing.T) {

	w := Woolworths{}
	err := w.Init(woolworthsServer.URL, ":memory:", PRODUCT_INFO_MAX_AGE)
	if err != nil {
		t.Fatal(err)
	}

	w.SaveDepartment("1-E5BEE36E")

	productIDs, err := w.GetProductList()
	if err != nil {
		t.Fatal(err)
	}
	if want, got := 38, len(productIDs); want != got {
		t.Errorf("Expected %d items, got %d", want, got)
	}

	var expectedProductIDs = []ProductID{133211, 134034, 105919}
	for i, id := range expectedProductIDs {
		if want, got := id, productIDs[i]; want != got {
			t.Errorf("Expected %d, got %d", want, got)
		}
	}
}