package coles

import (
	"fmt"
	"log/slog"
	"testing"
	"time"
)

func TestNewDepartmentInfoWorker(t *testing.T) {
	c := getInitialisedColes()
	go c.newDepartmentInfoWorker()
	// Wait for the worker to run
	time.Sleep(1 * time.Second)

	// Check that the department list has been updated
	departments, err := c.loadDepartmentInfoList()
	if err != nil {
		t.Fatalf("Failed to load department list: %v", err)
	}
	if want, got := 16, len(departments); want != got {
		t.Errorf("Expected %d departments, got %d", want, got)
	}

}

func TestDepartmentPageUpdateQueueWorker(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	departmentPageChannel := make(chan departmentPage)
	c := getInitialisedColes()
	// We want to get pages from this department, updated an hour ago.
	c.saveDepartment(departmentInfo{SeoToken: "1-E5BEE36E", Name: "Fruit & Vegetables", ProductCount: PRODUCTS_PER_PAGE * 3, Updated: time.Now().Add(-1 * time.Hour)})
	// We don't want to get pages from this department, updated an hour in the future.
	c.saveDepartment(departmentInfo{SeoToken: "1-E5BEE36F", Name: "Vruit & Fegetables", ProductCount: PRODUCTS_PER_PAGE * 3, Updated: time.Now().Add(1 * time.Hour)})
	c.listingPageUpdateInterval = 1 * time.Second
	go c.departmentPageUpdateQueueWorker(departmentPageChannel, 1*time.Second)

	pageIndex := 1
	for dp := range departmentPageChannel {
		if want, got := "1-E5BEE36E", dp.ID; want != got {
			t.Errorf("Expected %s, got %s", want, got)
		}
		if want, got := pageIndex, dp.page; want != got {
			t.Errorf("Expected %d, got %d", want, got)
		}
		fmt.Println("Got a page", dp)
		pageIndex++
		if pageIndex > 3 {
			break
		}
	}
	// Now test there are no more products waiting.
	select {
	case dept := <-departmentPageChannel:
		t.Fatal("Expected no more products, got", dept)
	case <-time.After(1 * time.Second):
	}
	// if want, got := 3, pageIndex; want != got {
	// 	t.Errorf("Expected %d, got %d", want, got)
	// }

}

func TestProductListPageWorker(t *testing.T) {
	c := getInitialisedColes()

	// Set up a department to scan products from
	dept := departmentInfo{SeoToken: "fruit-vegetables", Name: "Fruit & Vegetables", Updated: time.Now()}
	c.saveDepartment(dept)

	departmentPageChannel := make(chan departmentPage)
	go c.productListPageWorker(departmentPageChannel)

	departmentPageChannel <- departmentPage{
		ID:   "fruit-vegetables",
		page: 1,
	}
	// TODO remove this hardcoded sleep and use a loop in a goroutine with a channel for the output
	// as I've done before in another test.
	time.Sleep(50 * time.Millisecond)
	readInfo, err := c.loadProductInfo(productID("2511791"))
	if err != nil {
		t.Fatal(err)
	}
	if want, got := "Bananas Mini Pack", readInfo.Info.Name; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
}
