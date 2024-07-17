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

func TestScheduler(t *testing.T) {
	server := WoolworthsHTTPServer()

	w := Woolworths{}
	// w.Init(server.URL, ":memory:")
	w.Init(server.URL, "junk/delme.db3", 5*time.Second)
	cancel := make(chan struct{})
	go w.RunScheduler(cancel)
	time.Sleep(10 * time.Second)
	close(cancel)
	// TODO validate the DB contents
}
