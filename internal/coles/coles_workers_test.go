package coles

import (
	"fmt"
	"testing"
	"time"
)

func TestNewDepartmentInfoWorker(t *testing.T) {
	c := getInitialisedColes()
	go c.newDepartmentInfoWorker()
	// Wait for the worker to run
	time.Sleep(3 * time.Second)

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
	departmentPageChannel := make(chan departmentPage)
	c := getInitialisedColes()
	c.filterDepartments = false
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
	time.Sleep(2500 * time.Millisecond)
	readInfo, err := c.loadProductInfo(productID("2511791"))
	if err != nil {
		t.Fatal(err)
	}
	if want, got := "Bananas Mini Pack", readInfo.Info.Name; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
}

func ValidateProduct(t *testing.T, w *Coles, id productID, expectedName string) error {
	prod, err := w.loadProductInfo(id)
	if err != nil {
		return fmt.Errorf("Failed to get product ID %s: %v", id, err)
	}
	if prod.Info.Name != expectedName {
		t.Logf("Expected '%s', got '%s'", expectedName, prod.Info.Name)
		return fmt.Errorf("Expected '%s', got '%s'", expectedName, prod.Info.Name)
	}
	if want, got := "Fruit & Vegetables", prod.departmentDescription; want != got {
		t.Fatalf("Expected '%s', got '%s'", want, got)
		return fmt.Errorf("Expected '%s', got '%s'", want, got)
	}
	if want, got := "fruit-vegetables", prod.departmentID; want != got {
		t.Fatalf("Expected '%s', got '%s'", want, got)
		return fmt.Errorf("Expected '%s', got '%s'", want, got)
	}
	return nil
}
func TestScheduler(t *testing.T) {
	c := Coles{}
	c.Init(colesServer.URL, "delme.db3", 100*time.Second)
	c.listingPageUpdateInterval = 1 * time.Second
	cancel := make(chan struct{})
	go c.Run(cancel)

	done := make(chan struct{})
	go func() {
		for {
			err1 := ValidateProduct(t, &c, "2511791", "Bananas Mini Pack")
			err2 := ValidateProduct(t, &c, "5111654", "Pink Lady Apples Medium")
			if err1 == nil && err2 == nil {
				close(done)
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()

	select {
	case <-time.After(10 * time.Second):
		t.Fatal("Timed out waiting for scheduler to finish")
	case <-done:

	}

	close(cancel)
}
