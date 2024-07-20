package woolworths

import (
	"log/slog"
	"testing"
	"time"
)

func TestProductInfoFetchingWorker(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	w := Woolworths{}
	err := w.Init(woolworthsServer.URL, ":memory:", PRODUCT_INFO_MAX_AGE)
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

	// Give it a bogus product that doesn't exist in the mocked webserver.
	productsThatNeedAnUpdateChannel <- 999999

	// Ensure we get a blank product ID back
	select {
	case productInfo := <-productInfoChannel:
		if want, got := "", productInfo.Info.Name; want != got {
			t.Errorf("Expected %s, got %s", want, got)
		}
		// Check the updated time is within the last 2 seconds
		if time.Since(productInfo.Updated) > 2*time.Second {
			t.Errorf("Expected updated time to be within the last 2 seconds, got %v", time.Since(productInfo.Updated))
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timed out waiting for product info")
	}

}

func TestNewDepartmentIDWorker(t *testing.T) {
	w := Woolworths{}
	err := w.Init(woolworthsServer.URL, ":memory:", PRODUCT_INFO_MAX_AGE)
	if err != nil {
		t.Fatal(err)
	}

	// Pre-load one existing department ID to check we're only being notified of
	// new ones
	w.SaveDepartment("1-E5BEE36E")

	departmentIDChannel := make(chan DepartmentID)
	go w.NewDepartmentIDWorker(departmentIDChannel)
	var index int
	var departmentIDs = []DepartmentID{"1_DEB537E", "1_D5A2236", "1_6E4F4E4"}
	select {
	case d := <-departmentIDChannel:
		if want, got := departmentIDs[index], d; want != got {
			t.Errorf("Expected %s, got %s", want, got)
		}
		index++
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for department ID")
	}
}

func TestNewProductWorker(t *testing.T) {
	w := Woolworths{}
	err := w.Init(woolworthsServer.URL, ":memory:", PRODUCT_INFO_MAX_AGE)
	if err != nil {
		t.Fatal(err)
	}

	// Set up a department to scan products from
	w.SaveDepartment("1-E5BEE36E")

	productIDChannel := make(chan WoolworthsProductInfo)
	go w.NewProductWorker(productIDChannel)
	var index int
	var productIDs = []ProductID{133211, 134034, 105919}
	select {
	case p := <-productIDChannel:
		if want, got := productIDs[index], p.ID; want != got {
			t.Errorf("Expected %d, got %d", want, got)
		}
		index++
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for product info")
	}

}
