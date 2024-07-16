package woolworths

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
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

func ReadEntireFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return []byte{}, err
	}
	defer f.Close()

	// Read the contents of the file
	testData, err := io.ReadAll(f)
	if err != nil {
		return []byte{}, err
	}
	return testData, nil
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

			if responseData, err = ReadEntireFile(fmt.Sprintf("data/%d.json", productID)); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else if strings.HasPrefix(r.URL.Path, "/shop/browse/fruit-veg") {

		} else if strings.HasPrefix(r.URL.Path, "/apis/ui/browse/category") {
			if responseData, err = ReadEntireFile("data/category.json"); err != nil {
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
	if want, got := 133211, prodIDs[0].ID; want != got {
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

		productInfo, err := w.GetProductInfo(ProductID{ID: id})
		if err != nil {
			t.Fatal(err)
		}

		if productInfo.Name != want {
			t.Errorf("Expected %s, got %s", want, productInfo.Name)
		}
	}
}
