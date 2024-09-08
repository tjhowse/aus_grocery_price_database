package woolworths

import (
	"fmt"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func ValidateProduct(t *testing.T, w *Woolworths, id productID, expectedName string) error {
	prod, err := w.loadProductInfo(id)
	if err != nil {
		return fmt.Errorf("Failed to get product ID %s: %v", id, err)
	}
	if prod.Info.DisplayName != expectedName {
		t.Logf("Expected '%s', got '%s'", expectedName, prod.Info.DisplayName)
		return fmt.Errorf("Expected '%s', got '%s'", expectedName, prod.Info.DisplayName)
	}
	if want, got := "Fruit & Veg", prod.departmentDescription; want != got {
		t.Fatalf("Expected '%s', got '%s'", want, got)
		return fmt.Errorf("Expected '%s', got '%s'", want, got)
	}
	if want, got := departmentID("1-E5BEE36E"), prod.departmentID; want != got {
		t.Fatalf("Expected '%s', got '%s'", want, got)
		return fmt.Errorf("Expected '%s', got '%s'", want, got)
	}
	return nil
}

func TestScheduler(t *testing.T) {
	w := Woolworths{}
	w.Init(woolworthsServer.URL, ":memory:", 100*time.Second)
	w.client.Ratelimiter = rate.NewLimiter(rate.Every(1*time.Millisecond), 1)
	w.listingPageUpdateInterval = 1 * time.Second
	w.filteredDepartmentIDsSet = map[departmentID]bool{
		"1-E5BEE36E": true, // Fruit & Veg
		"1_DEB537E":  true, // Bakery
	}
	w.filterDepartments = true
	cancel := make(chan struct{})
	go w.Run(cancel)

	done := make(chan struct{})
	go func() {
		for {
			err1 := ValidateProduct(t, &w, "165262", "Raspberries 125g Punnet")
			err2 := ValidateProduct(t, &w, "187314", "Woolworths Broccolini Bunch Each")
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
