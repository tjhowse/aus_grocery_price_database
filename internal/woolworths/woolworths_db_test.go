package woolworths

import (
	"encoding/json"
	"testing"

	utils "github.com/tjhowse/aus_grocery_price_database/internal/utils"
)

func TestUpdateProductInfo(t *testing.T) {
	var err error
	var prodInfoRaw []byte
	prodInfoRaw, err = utils.ReadEntireFile("data/187314.json")
	if err != nil {
		t.Fatal(err)
	}
	prodInfo := ProductInfo{}
	err = json.Unmarshal(prodInfoRaw, &prodInfo)
	if err != nil {
		t.Fatal(err)
	}

	w := Woolworths{}
	// w.Init("https://www.woolworths.com.au", ":memory:")
	w.Init("https://www.woolworths.com.au", "delme.db3")
	err = w.SaveProductInfo(prodInfo)
	if err != nil {
		t.Fatal(err)
	}

	var readProdInfo ProductInfo
	readProdInfo, err = w.LoadProductInfo(187314)
	if err != nil {
		t.Fatal(err)
	}
	if readProdInfo.Description != prodInfo.Description {
		t.Errorf("Expected %v, got %v", prodInfo.Description, readProdInfo.Description)
	}
}
