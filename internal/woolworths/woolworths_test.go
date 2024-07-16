package woolworths

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

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
			if responseData, err = utils.ReadEntireFile("data/category.json"); err != nil {
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

func TestGetProductList(t *testing.T) {
	server := WoolworthsHTTPServer()
	defer server.Close()

	w := Woolworths{}
	w.Init(server.URL)

	prodIDs, err := w.GetProductList()
	if err != nil {
		t.Fatal(err)
	}
	if want, got := ProductID(133211), prodIDs[0]; want != got {
		t.Errorf("Expected %d, got %d", want, got)
	}
}

func TestGetProductInfo(t *testing.T) {
	server := WoolworthsHTTPServer()
	defer server.Close()

	w := Woolworths{}
	w.Init(server.URL)

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

		if productInfo.Name != want {
			t.Errorf("Expected %s, got %s", want, productInfo.Name)
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
	w.Init(server.URL)

	departmentIDs, err := w.GetDepartmentIDs()
	if err != nil {
		t.Fatal(err)
	}
	if want, got := DepartmentID("1-E5BEE36E"), departmentIDs[0]; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
}
