package main

import (
	"log/slog"
	"strconv"
	"testing"
	"time"

	shared "github.com/tjhowse/aus_grocery_price_database/internal/shared"
)

type MockInfluxDB struct {
	url, token, org, bucket          string
	writtenProductDataPoints         []shared.ProductInfo
	writtenArbitrarySystemDatapoints []struct {
		field string
		value interface{}
	}
	writtenSystemDatapoints []SystemStatusDatapoint
	closed                  bool
}

func (i *MockInfluxDB) Init(url, token, org, bucket string) {
	i.url = url
	i.token = token
	i.org = org
	i.bucket = bucket
	i.closed = false
}

func (i *MockInfluxDB) WriteProductDatapoint(info shared.ProductInfo) {
	i.writtenProductDataPoints = append(i.writtenProductDataPoints, info)
	slog.Info("Writing product datapoint", "name", info.Name, "store", info.Store, "location", info.Location, "department", info.Department, "cents", info.PriceCents, "grams", info.WeightGrams)
}

func (i *MockInfluxDB) WriteArbitrarySystemDatapoint(field string, value interface{}) {
	i.writtenArbitrarySystemDatapoints = append(i.writtenArbitrarySystemDatapoints, struct {
		field string
		value interface{}
	}{field, value})
}

func (i *MockInfluxDB) WriteSystemDatapoint(data SystemStatusDatapoint) {
	i.writtenSystemDatapoints = append(i.writtenSystemDatapoints, data)
}

func (i *MockInfluxDB) WriteWorker(input <-chan shared.ProductInfo) {
	for info := range input {
		i.WriteProductDatapoint(info)
	}
}

func (i *MockInfluxDB) Close() {
	i.closed = true
}

type MockGroceryStore struct {
	url, dbpath   string
	productMaxAge time.Duration
}

func (m *MockGroceryStore) Init(url string, dbpath string, age time.Duration) error {
	m.url = url
	m.dbpath = dbpath
	m.productMaxAge = age
	return nil
}

func (m *MockGroceryStore) Run(chan struct{}) {
}

func (m *MockGroceryStore) GetSharedProductsUpdatedAfter(cutoff time.Time, count int) ([]shared.ProductInfo, error) {
	var productIDs []shared.ProductInfo

	for i := 0; i < count; i++ {

		productIDs = append(productIDs, shared.ProductInfo{
			ID:          strconv.Itoa(i),
			Name:        "Test Product" + strconv.Itoa(i),
			Description: "Test Description" + strconv.Itoa(i),
			Store:       "Test Store" + strconv.Itoa(i),
			Department:  "Test Department" + strconv.Itoa(i),
			Location:    "Test Location" + strconv.Itoa(i),
			PriceCents:  100 + i,
			WeightGrams: 1000 + i,
			Timestamp:   time.Now().Add(-5 * time.Minute),
		})
	}

	return productIDs, nil
}

func (m *MockGroceryStore) GetTotalProductCount() (int, error) {
	return 100, nil
}

func TestRun(t *testing.T) {
	mockGroceryStore := MockGroceryStore{}
	mockInfluxDB := MockInfluxDB{}
	config := config{

		InfluxDBURL:                 "a",
		InfluxDBToken:               "b",
		InfluxDBOrg:                 "c",
		InfluxDBBucket:              "d",
		InfluxUpdateIntervalSeconds: 1,
		LocalWoolworthsDBPath:       "e",
		MaxProductAgeMinutes:        1,
		WoolworthsURL:               "f",
		DebugLogging:                false,
	}

	running := true

	go run(&running, &config, &mockInfluxDB, &mockGroceryStore)

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if len(mockInfluxDB.writtenProductDataPoints) == 100 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	running = false

	if want, got := 100, len(mockInfluxDB.writtenProductDataPoints); want != got {
		t.Fatalf("Expected %d products, got %d", want, got)
	}

	if want, got := "Test Product0", mockInfluxDB.writtenProductDataPoints[0].Name; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}

	if want, got := "Test Department0", mockInfluxDB.writtenProductDataPoints[0].Department; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}

	if want, got := 100, mockInfluxDB.writtenProductDataPoints[0].PriceCents; want != got {
		t.Errorf("Expected %d, got %d", want, got)
	}
	// Give time for the timeseries database to be closed down
	time.Sleep(2 * time.Second)
	if !mockInfluxDB.closed {
		t.Error("Expected the database to be closed")
	}

	// Validate the db and grocery were initialised correctly
	if want, got := "f", mockGroceryStore.url; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}

	if want, got := "e", mockGroceryStore.dbpath; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}

	if want, got := 1*time.Minute, mockGroceryStore.productMaxAge; want != got {
		t.Errorf("Expected %v, got %v", want, got)
	}

	if want, got := "a", mockInfluxDB.url; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}

	if want, got := "b", mockInfluxDB.token; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}

	if want, got := "c", mockInfluxDB.org; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}

	if want, got := "d", mockInfluxDB.bucket; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}

	if want, got := 1, len(mockInfluxDB.writtenArbitrarySystemDatapoints); want != got {
		t.Fatalf("Expected %d items, got %d", want, got)
	}

	if want, got := 100, mockInfluxDB.writtenSystemDatapoints[0].TotalProductCount; want != got {
		t.Errorf("Expected %v, got %v", want, got)
	}

}
