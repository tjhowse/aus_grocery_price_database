package woolworths

import (
	"fmt"
	"log/slog"
	"testing"
	"time"
)

func ValidateProduct(t *testing.T, w *Woolworths, id productID, want string) error {
	prod, err := w.loadProductInfo(id)
	if err != nil {
		return fmt.Errorf("Failed to get product ID %s: %v", id, err)
	}
	if prod.Info.Name != want {
		return fmt.Errorf("Expected %s, got %s", want, prod.Info.Name)
	}
	return nil
}

func TestScheduler(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	w := Woolworths{}
	w.Init(woolworthsServer.URL, ":memory:", 100*time.Second)
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
			err1 := ValidateProduct(t, &w, "165262", "Driscoll's Raspberries Punnet 125g Punnet")
			err2 := ValidateProduct(t, &w, "187314", "Woolworths Broccolini Bunch  Each")
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

// func TestSchedulerExtended(t *testing.T) {
// 	slog.SetLogLoggerLevel(slog.LevelDebug)

// 	w := Woolworths{}
// 	w.Init(woolworthsServer.URL, ":memory:", 100*time.Second)
// 	w.filteredDepartmentIDsSet = map[departmentID]bool{
// 		"1-E5BEE36E": true, // Fruit & Veg
// 		"1_DEB537E":  true, // Bakery
// 	}
// 	w.filterDepartments = true
// 	cancel := make(chan struct{})
// 	go w.RunExtended(cancel)

// 	done := make(chan struct{})
// 	go func() {
// 		for {
// 			err1 := ValidateProduct(t, &w, "165262", "Driscoll's Raspberries Punnet 125g Punnet")
// 			err2 := ValidateProduct(t, &w, "187314", "Woolworths Broccolini Bunch  Each")
// 			if err1 == nil && err2 == nil {
// 				close(done)
// 				return
// 			}
// 			time.Sleep(1 * time.Second)
// 		}
// 	}()

// 	select {
// 	case <-time.After(10 * time.Second):
// 		t.Fatal("Timed out waiting for scheduler to finish")
// 	case <-done:

// 	}

// 	close(cancel)
// }
