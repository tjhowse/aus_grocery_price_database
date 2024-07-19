package woolworths

import (
	"encoding/json"
	"log/slog"
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
	readProdInfo, err = w.LoadProductInfo(187314)
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

	result = WoolworthsProductInfo{ID: 187314, Info: prodInfo}
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
		if id != 187314 {
			t.Errorf("Expected 187314, got %d", id)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for product ID")
	}
}

func TestMissingProduct(t *testing.T) {

	server := WoolworthsHTTPServer()

	w := Woolworths{}
	w.Init(server.URL, ":memory:", 5*time.Second)
	_, err := w.LoadProductInfo(123456)
	if err == nil {
		t.Fatal("Expected an error")
	}
	if want, got := ErrProductMissing, err; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
}

func ValidateProduct(t *testing.T, w *Woolworths, id ProductID, want string) {
	prod, err := w.LoadProductInfo(id)
	if err != nil {
		t.Fatal(err)
	}
	if prod.Name != want {
		t.Errorf("Expected %s, got %s", want, prod.Name)
	}
}

func TestScheduler(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	server := WoolworthsHTTPServer()

	w := Woolworths{}
	w.Init(server.URL, ":memory:", 5*time.Second)
	cancel := make(chan struct{})
	go w.RunScheduler(cancel)
	time.Sleep(5 * time.Second)
	close(cancel)
	ValidateProduct(t, &w, 165262, "Driscoll's Raspberries Punnet 125g Punnet")
	ValidateProduct(t, &w, 187314, "Woolworths Broccolini Bunch  Each")
	ValidateProduct(t, &w, 524336, "Woolworths Baby Spinach Spinach 280g")
}
