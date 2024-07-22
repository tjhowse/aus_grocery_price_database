package woolworths

import (
	"encoding/json"
	"testing"
	"time"

	utils "github.com/tjhowse/aus_grocery_price_database/internal/utils"
)

func TestUpdateProductInfo(t *testing.T) {
	wProdInfo, err := ReadWoolworthsProductInfoFromFile("data/187314.json")
	if err != nil {
		t.Fatal(err)
	}

	w := Woolworths{}
	w.Init("https://www.woolworths.com.au", ":memory:", PRODUCT_INFO_MAX_AGE)

	wProdInfo.Updated = time.Now()
	err = w.SaveProductInfo(wProdInfo)
	if err != nil {
		t.Fatal(err)
	}

	var readProdInfo ProductInfo
	readProdInfo, err = w.LoadProductInfo("187314")
	if err != nil {
		t.Fatal(err)
	}
	if readProdInfo.Description != wProdInfo.Info.Description {
		t.Errorf("Expected %v, got %v", wProdInfo.Info.Description, readProdInfo.Description)
	}
}

func ReadWoolworthsProductInfoFromFile(filename string) (WoolworthsProductInfo, error) {
	var err error
	var prodInfoRaw []byte
	var result WoolworthsProductInfo
	prodInfoRaw, err = utils.ReadEntireFile(filename)
	if err != nil {
		return result, err
	}
	prodInfo := ProductInfo{}
	err = json.Unmarshal(prodInfoRaw, &prodInfo)
	if err != nil {
		return result, err
	}

	result = WoolworthsProductInfo{ID: "187314", Info: prodInfo, RawJSON: prodInfoRaw}
	return result, nil
}

func TestProductUpdateQueueGenerator(t *testing.T) {
	wProdInfo, err := ReadWoolworthsProductInfoFromFile("data/187314.json")
	if err != nil {
		t.Fatal(err)
	}

	w := Woolworths{}
	w.Init("https://www.woolworths.example.com", ":memory:", PRODUCT_INFO_MAX_AGE)

	wProdInfo.Updated = time.Now().Add(-1 * time.Hour)
	err = w.SaveProductInfo(wProdInfo)
	if err != nil {
		t.Fatal(err)
	}

	idChannel := make(chan ProductID)
	go w.ProductUpdateQueueWorker(idChannel, 20*time.Millisecond)

	time.Sleep(50 * time.Microsecond)

	select {
	case id := <-idChannel:
		if id != "187314" {
			t.Errorf("Expected 187314, got %s", id)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for product ID")
	}
}

func TestMissingProduct(t *testing.T) {
	w := Woolworths{}
	w.Init(woolworthsServer.URL, ":memory:", 5*time.Second)
	_, err := w.LoadProductInfo("123456")
	if err == nil {
		t.Fatal("Expected an error")
	}
	if want, got := ErrProductMissing, err; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
}

func TestDepartment(t *testing.T) {
	w := Woolworths{}
	w.Init(woolworthsServer.URL, ":memory:", 5*time.Second)
	w.SaveDepartment("1-E5BEE36E")
	departmentIDs, err := w.LoadDepartmentIDsList()
	if err != nil {
		t.Fatal(err)
	}
	if want, got := 1, len(departmentIDs); want != got {
		t.Errorf("Expected %d, got %d", want, got)
	}
	if want, got := DepartmentID("1-E5BEE36E"), departmentIDs[0]; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
}

func TestDBFail(t *testing.T) {

	w := Woolworths{}
	err := w.Init("", "/zingabingo/db.db3", 5*time.Second)
	if err == nil {
		t.Fatal("Expected an error")
	}
	if want, got := "failed to create blank DB: unable to open database file: no such file or directory", err.Error(); want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
}

func TestGetProductIDsUpdatedAfter(t *testing.T) {
	w := Woolworths{}
	w.Init(woolworthsServer.URL, ":memory:", 5*time.Second)
	w.SaveProductInfo(WoolworthsProductInfo{ID: "123455", Info: ProductInfo{}, Updated: time.Now().Add(-5 * time.Minute)})
	w.SaveProductInfo(WoolworthsProductInfo{ID: "123456", Info: ProductInfo{}, Updated: time.Now().Add(-4 * time.Minute)})
	w.SaveProductInfo(WoolworthsProductInfo{ID: "123457", Info: ProductInfo{}, Updated: time.Now().Add(-3 * time.Minute)})
	w.SaveProductInfo(WoolworthsProductInfo{ID: "123458", Info: ProductInfo{}, Updated: time.Now().Add(-1 * time.Minute)})
	w.SaveProductInfo(WoolworthsProductInfo{ID: "123459", Info: ProductInfo{}, Updated: time.Now()})
	productIDs, err := w.GetProductIDsUpdatedAfter(time.Now().Add(-2*time.Minute), 10)
	if err != nil {
		t.Fatal(err)
	}
	if want, got := 2, len(productIDs); want != got {
		t.Errorf("Expected %d products, got %d", want, got)
	}
	if want, got := ProductID("123458"), productIDs[0]; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
	if want, got := ProductID("123459"), productIDs[1]; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
	productIDs, err = w.GetProductIDsUpdatedAfter(time.Now().Add(-2*time.Minute), 1)
	if err != nil {
		t.Fatal(err)
	}
	if want, got := 1, len(productIDs); want != got {
		t.Errorf("Expected %d products, got %d", want, got)
	}

}
