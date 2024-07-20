package woolworths

import (
	"bufio"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

func TestProductInfoFetchingWorker(t *testing.T) {
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

	// Set up a pipe to use as a substitute for stdout for the logger
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	defer writer.Close()

	// Tell slog to log to the pipe instead of stdout
	slog.SetDefault(slog.New(slog.NewTextHandler(writer, &slog.HandlerOptions{Level: slog.LevelDebug})))
	// TODO Defer a func call to restore slog settings back to how they were before this test.

	// Give it a bogus product that doesn't exist in the mocked webserver.
	productsThatNeedAnUpdateChannel <- 999999

	// Ensure we don't get a productInfo from the worker.
	select {
	case productInfo := <-productInfoChannel:
		t.Fatalf("Expected nothing, got %v", productInfo)
	case <-time.After(100 * time.Millisecond):
	}

	// Set up a semaphore for the goroutine to signal to us that it found the log message
	foundLogMessage := make(chan struct{})

	// Run a goroutine to scan the output of the logger for the expected message
	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "failed to get product info") {
				foundLogMessage <- struct{}{}
			}
		}
	}()

	// Wait for the log message to be found, or timeout.
	select {
	case <-foundLogMessage:
	case <-time.After(1000 * time.Millisecond):
		t.Fatal("Timed out waiting for log message")
	}

	// If this test seems tortuous, I agree. I'd like an easier way to capture log output from tests.
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
