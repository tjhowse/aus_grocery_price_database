package influxdb

import (
	"context"
	"os"
	"testing"
	"time"

	shared "github.com/tjhowse/aus_grocery_price_database/internal/shared"
)

// TestPriceHistoryIntegration runs against a real, empty InfluxDB 2 bucket. E.G.:
// docker run -p 8086:8086 -e DOCKER_INFLUXDB_INIT_MODE=setup -e DOCKER_INFLUXDB_INIT_USERNAME=admin \
// -e DOCKER_INFLUXDB_INIT_PASSWORD=adminpass -e DOCKER_INFLUXDB_INIT_ORG=groceries \
// -e DOCKER_INFLUXDB_INIT_BUCKET=groceries -e DOCKER_INFLUXDB_INIT_ADMIN_TOKEN=testtoken influxdb:2.7.7
// INFLUXDB_TEST_URL=http://localhost:8086 INFLUXDB_TEST_TOKEN=testtoken go test ./internal/databases/influxdb/
func TestPriceHistoryIntegration(t *testing.T) {
	url, token := os.Getenv("INFLUXDB_TEST_URL"), os.Getenv("INFLUXDB_TEST_TOKEN")
	if url == "" || token == "" {
		t.Skip("INFLUXDB_TEST_URL and INFLUXDB_TEST_TOKEN not set")
	}
	i := InfluxDB{}
	i.Init(url, token, "groceries", "groceries")
	i.InitQuery(url, token, "groceries")
	defer i.Close()

	id := "woolworths_sku_" + time.Now().Format("150405.000000")
	day := func(d, hour int) time.Time { return time.Date(2025, 1, d, hour, 0, 0, 0, time.UTC) }
	for _, p := range []shared.ProductInfo{
		{ID: id, Name: "Oranges", PriceCents: 500, Timestamp: day(1, 1)},
		// The last price of the day wins.
		{ID: id, Name: "Oranges", PriceCents: 450, Timestamp: day(1, 20)},
		// A rename is a new series, which should be merged with the old one.
		{ID: id, Name: "Navel Oranges", PriceCents: 400, Timestamp: day(3, 2)},
		{ID: "some_other_product", Name: "Apples", PriceCents: 100, Timestamp: day(2, 2)},
	} {
		i.WriteProductDatapoint(p)
	}
	i.groceryWriteAPI.Flush()

	history, err := i.GetPriceHistory(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	want := []shared.PricePoint{
		{Time: day(1, 0), PriceCents: 450},
		{Time: day(3, 0), PriceCents: 400},
	}
	if len(history) != len(want) {
		t.Fatalf("Expected %v, got %v", want, history)
	}
	for n := range want {
		if !want[n].Time.Equal(history[n].Time) || want[n].PriceCents != history[n].PriceCents {
			t.Errorf("Expected %v, got %v", want[n], history[n])
		}
	}
}
