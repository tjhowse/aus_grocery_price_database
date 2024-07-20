package woolworths

import (
	"log/slog"
	"testing"
	"time"
)

func TestScheduler(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	w := Woolworths{}
	w.Init(woolworthsServer.URL, ":memory:", 10*time.Second)
	cancel := make(chan struct{})
	go w.RunScheduler(cancel)
	time.Sleep(5 * time.Second)
	close(cancel)
	// TODO put these tests in place
	// ValidateProduct(t, &w, 165262, "Driscoll's Raspberries Punnet 125g Punnet")
	// ValidateProduct(t, &w, 187314, "Woolworths Broccolini Bunch  Each")
	// ValidateProduct(t, &w, 524336, "Woolworths Baby Spinach Spinach 280g")
}
