package woolworths

import (
	"fmt"
	"log/slog"
	"testing"
	"time"
)

func ValidateProduct(t *testing.T, w *Woolworths, id ProductID, want string) error {
	prod, err := w.LoadProductInfo(id)
	if err != nil {
		return fmt.Errorf("Failed to get product ID %s: %v", id, err)
	}
	if prod.Name != want {
		return fmt.Errorf("Expected %s, got %s", want, prod.Name)
	}
	return nil
}

func TestScheduler(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	w := Woolworths{}
	w.Init(woolworthsServer.URL, ":memory:", 100*time.Second)
	cancel := make(chan struct{})
	go w.RunScheduler(cancel)

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
