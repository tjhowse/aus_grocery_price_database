package woolworths

import (
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	utils "github.com/tjhowse/aus_grocery_price_database/internal/utils"
)

func TestUpdateProductInfo(t *testing.T) {
	wProdInfo, err := ReadWoolworthsProductInfoFromFile("data/187314.json")
	if err != nil {
		t.Fatal(err)
	}

	w := getInitialisedWoolworths()

	wProdInfo.Updated = time.Now()
	err = w.saveProductInfo(wProdInfo)
	if err != nil {
		t.Fatal(err)
	}

	var readProdInfo woolworthsProductInfo
	readProdInfo, err = w.loadProductInfo("187314")
	if err != nil {
		t.Fatal(err)
	}
	if readProdInfo.Info.Description != wProdInfo.Info.Description {
		t.Errorf("Expected %v, got %v", wProdInfo.Info.Description, readProdInfo.Info.Description)
	}
}

func ReadWoolworthsProductInfoFromFile(filename string) (woolworthsProductInfo, error) {
	var err error
	var prodInfoRaw []byte
	var result woolworthsProductInfo
	prodInfoRaw, err = utils.ReadEntireFile(filename)
	if err != nil {
		return result, err
	}
	prodInfo := productInfo{}
	err = json.Unmarshal(prodInfoRaw, &prodInfo)
	if err != nil {
		return result, err
	}

	result = woolworthsProductInfo{ID: productID(prodInfo.Sku), Info: prodInfo, RawJSON: prodInfoRaw}
	return result, nil
}

func TestProductUpdateQueueGenerator(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	var err error
	var wProdInfo1, wProdInfo2 woolworthsProductInfo
	wProdInfo1, err = ReadWoolworthsProductInfoFromFile("data/187314.json")
	if err != nil {
		t.Fatal(err)
	}

	wProdInfo2, err = ReadWoolworthsProductInfoFromFile("data/524336.json")
	if err != nil {
		t.Fatal(err)
	}
	w := getInitialisedWoolworths()

	wProdInfo1.Updated = time.Now().Add(-2 * time.Hour)
	wProdInfo2.Updated = time.Now().Add(-1 * time.Hour)

	err = w.saveProductInfo(wProdInfo1)
	if err != nil {
		t.Fatal(err)
	}
	err = w.saveProductInfo(wProdInfo2)
	if err != nil {
		t.Fatal(err)
	}

	idChannel := make(chan productID)
	go w.productUpdateQueueWorker(idChannel, 20*time.Millisecond, 1)

	time.Sleep(50 * time.Microsecond)

	select {
	case id := <-idChannel:
		if id != "187314" {
			t.Errorf("Expected 187314, got %s", id)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for first product ID")
	}

	// Ensure we don't get the same product twice.
	time.Sleep(50 * time.Microsecond)

	select {
	case id := <-idChannel:
		if id != "524336" {
			t.Errorf("Expected 524336, got %s", id)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for second product ID")
	}
}

func TestMissingProduct(t *testing.T) {
	w := getInitialisedWoolworths()
	_, err := w.loadProductInfo("123456")
	if err == nil {
		t.Fatal("Expected an error")
	}
	if want, got := ErrProductMissing, err; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
}

func TestDepartmentInfo(t *testing.T) {
	w := getInitialisedWoolworths()
	dept := departmentInfo{NodeID: "1-E5BEE36E", Description: "Fruit & Veg", Updated: time.Now()}
	w.saveDepartment(dept)
	departmentIDs, err := w.loadDepartmentInfoList()
	if err != nil {
		t.Fatal(err)
	}
	if want, got := 1, len(departmentIDs); want != got {
		t.Errorf("Expected %d, got %d", want, got)
	}
	if want, got := departmentID("1-E5BEE36E"), departmentIDs[0].NodeID; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
	if want, got := "Fruit & Veg", departmentIDs[0].Description; want != got {
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

func TestGetSharedProductsUpdatedAfter(t *testing.T) {
	w := Woolworths{}
	w.Init(woolworthsServer.URL, ":memory:", 5*time.Second)
	w.filterDepartments = false
	var infoList []woolworthsProductInfo
	infoList = append(infoList, woolworthsProductInfo{ID: "123455", Info: productInfo{Name: "1", Offers: offer{Price: decimal.NewFromFloat(1.5)}}, Updated: time.Now().Add(-5 * time.Minute)})
	infoList = append(infoList, woolworthsProductInfo{ID: "123456", Info: productInfo{Name: "1", Offers: offer{Price: decimal.NewFromFloat(2.4)}}, Updated: time.Now().Add(-4 * time.Minute)})
	infoList = append(infoList, woolworthsProductInfo{ID: "123457", Info: productInfo{Name: "1", Offers: offer{Price: decimal.NewFromFloat(3.3)}}, Updated: time.Now().Add(-3 * time.Minute)})
	infoList = append(infoList, woolworthsProductInfo{ID: "123458", Info: productInfo{Name: "1", Offers: offer{Price: decimal.NewFromFloat(4.2)}}, Updated: time.Now().Add(-1 * time.Minute)})
	infoList = append(infoList, woolworthsProductInfo{ID: "123459", Info: productInfo{Name: "1", Offers: offer{Price: decimal.NewFromFloat(5.1)}}, Updated: time.Now()})
	infoList = append(infoList, woolworthsProductInfo{ID: "123460", Info: productInfo{Offers: offer{Price: decimal.NewFromFloat(6.0)}}, Updated: time.Now()})

	for _, info := range infoList {
		w.saveProductInfo(info)
	}
	productIDs, err := w.GetSharedProductsUpdatedAfter(time.Now().Add(-2*time.Minute), 10)
	if err != nil {
		t.Fatal(err)
	}
	if want, got := 2, len(productIDs); want != got {
		t.Fatalf("Expected %d products, got %d", want, got)
	}
	if want, got := WOOLWORTHS_ID_PREFIX+"123458", productIDs[0].ID; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
	if want, got := WOOLWORTHS_ID_PREFIX+"123459", productIDs[1].ID; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
	if want, got := 510, productIDs[1].PriceCents; want != got {
		t.Errorf("Expected %v, got %v", want, got)
	}
	productIDs, err = w.GetSharedProductsUpdatedAfter(time.Now().Add(-2*time.Minute), 1)
	if err != nil {
		t.Fatal(err)
	}
	if want, got := 1, len(productIDs); want != got {
		t.Fatalf("Expected %d products, got %d", want, got)
	}
	if want, got := WOOLWORTHS_ID_PREFIX+"123458", productIDs[0].ID; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
	if want, got := 420, productIDs[0].PriceCents; want != got {
		t.Errorf("Expected %v, got %v", want, got)
	}

	if total, err := w.GetTotalProductCount(); err != nil {
		t.Fatal(err)
	} else {
		if want, got := 6, total; want != got {
			t.Errorf("Expected %d, got %d", want, got)
		}
	}

}

func TestCheckIfKnownProductID(t *testing.T) {
	w := getInitialisedWoolworths()
	w.saveProductInfo(woolworthsProductInfo{ID: "123456", Info: productInfo{Name: "1", Offers: offer{Price: decimal.NewFromFloat(1.5)}}, Updated: time.Now().Add(-5 * time.Minute)})
	found, err := w.checkIfKnownProductID("123456")
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatal("Didn't find a product as expected")
	}
	found, err = w.checkIfKnownProductID("123457")
	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Fatal("Found a product we weren't expecting to find.")
	}
}

func TestSaveProductInfo(t *testing.T) {
	w := getInitialisedWoolworths()
	inProduct := woolworthsProductInfo{ID: "123456", departmentID: "abc", departmentDescription: "cba", Info: productInfo{Name: "1", Offers: offer{Price: decimal.NewFromFloat(1.5)}}, Updated: time.Now()}

	err := w.saveProductInfo(inProduct)
	if err != nil {
		t.Fatal(err)
	}
	outProduct, err := w.loadProductInfo("123456")
	if err != nil {
		t.Fatal(err)
	}
	if want, got := inProduct.Info.Name, outProduct.Info.Name; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
	if want, got := inProduct.Info.Offers.Price.Mul(decimal.NewFromInt(100)), outProduct.Info.Offers.Price; want.Cmp(got) != 0 {
		t.Errorf("Expected %v, got %v", want, got)
	}
	if want, got := inProduct.departmentID, outProduct.departmentID; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
	if want, got := inProduct.departmentDescription, outProduct.departmentDescription; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}

}
