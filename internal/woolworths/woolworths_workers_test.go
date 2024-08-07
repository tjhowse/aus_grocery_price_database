package woolworths

import (
	"log/slog"
	"testing"
	"time"
)

func TestProductInfoFetchingWorker(t *testing.T) {
	t.Fatal("Confirm a failed test blocks deployment")
	slog.SetLogLoggerLevel(slog.LevelDebug)
	w := getInitialisedWoolworths()

	productInfoChannel := make(chan woolworthsProductInfo)
	productsThatNeedAnUpdateChannel := make(chan productID)
	go w.productInfoFetchingWorker(productsThatNeedAnUpdateChannel, productInfoChannel)

	productsThatNeedAnUpdateChannel <- "187314"

	select {
	case productInfo := <-productInfoChannel:
		if want, got := "Woolworths Broccolini Bunch  Each", productInfo.Info.Name; want != got {
			t.Errorf("Expected %s, got %s", want, got)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timed out waiting for product info")
	}

	// Give it a bogus product that doesn't exist in the mocked webserver.
	productsThatNeedAnUpdateChannel <- "999999"

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
	w := getInitialisedWoolworths()

	// Pre-load one existing department ID to check we're only being notified of
	// new ones
	dept := departmentInfo{NodeID: "1-E5BEE36E", Description: "Fruit & Vegetables"}
	w.saveDepartment(dept)

	departmentIDChannel := make(chan departmentInfo)
	go w.newDepartmentInfoWorker(departmentIDChannel)
	var index int
	// var departmentIDs = []departmentID{"1_DEB537E", "1_D5A2236", "1_6E4F4E4"}
	var departmentIDs = []departmentID{"specialsgroup", "1_DEF0CCD", "1_D5A2236"}
	select {
	case d := <-departmentIDChannel:
		if want, got := departmentIDs[index], d.NodeID; want != got {
			t.Errorf("Expected %s, got %s", want, got)
		}
		index++
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for department ID")
	}
}

func TestNewProductWorker(t *testing.T) {
	w := getInitialisedWoolworths()

	// Set up a department to scan products from
	dept := departmentInfo{NodeID: "1-E5BEE36E", Description: "Fruit & Vegetables"}
	w.saveDepartment(dept)

	productIDChannel := make(chan woolworthsProductInfo)
	go w.newProductWorker(productIDChannel)
	var index int
	var productIDs = []productID{"133211", "134034", "105919"}
	select {
	case p := <-productIDChannel:
		if want, got := productIDs[index], p.ID; want != got {
			t.Errorf("Expected %s, got %s", want, got)
		}
		if want, got := "Fruit & Vegetables", p.departmentDescription; want != got {
			t.Errorf("Expected %s, got %s", want, got)
		}
		index++
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for product info")
	}

}
