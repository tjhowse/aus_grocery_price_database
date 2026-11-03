package influxdb

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	shared "github.com/tjhowse/aus_grocery_price_database/internal/shared"
)

type InfluxDB struct {
	db              influxdb2.Client
	groceryWriteAPI api.WriteAPI
	systemWriteAPI  api.WriteAPI
	queryDB         influxdb2.Client
	queryAPI        api.QueryAPI
	bucket          string
}

func (i *InfluxDB) Init(url, token, org, bucket string) {
	slog.Info("Initialising InfluxDB", "url", url, "org", org, "bucket", bucket)
	i.db = influxdb2.NewClient(url, token)
	i.groceryWriteAPI = i.db.WriteAPI(org, bucket)
	i.systemWriteAPI = i.db.WriteAPI(org, "system")
	i.bucket = bucket
}

// InitQuery sets up reading price history. The write token is write-only,
// so this takes a separate token with read access to the bucket.
func (i *InfluxDB) InitQuery(url, token, org string) {
	slog.Info("Initialising InfluxDB queries", "url", url, "org", org, "bucket", i.bucket)
	i.queryDB = influxdb2.NewClient(url, token)
	i.queryAPI = i.queryDB.QueryAPI(org)
}

// escapeFluxString makes s safe to embed in a double-quoted Flux string literal.
func escapeFluxString(s string) string {
	return strings.NewReplacer(`\`, `\\`, `"`, `\"`, "$", `\$`).Replace(s)
}

// priceHistoryQuery returns a Flux query for the last price each day (UTC) of the product
// with the given shared ID, oldest first. group() merges series split by a renamed product.
func priceHistoryQuery(bucket, id string) string {
	return fmt.Sprintf(`from(bucket: "%s")
	|> range(start: 0)
	|> filter(fn: (r) => r._measurement == "product" and r._field == "cents" and r.id == "%s")
	|> group()
	|> sort(columns: ["_time"])
	|> aggregateWindow(every: 1d, fn: last, createEmpty: false, timeSrc: "_start")`,
		escapeFluxString(bucket), escapeFluxString(id))
}

// GetPriceHistory returns the daily price of the product with the given shared ID, oldest first.
func (i *InfluxDB) GetPriceHistory(ctx context.Context, id string) ([]shared.PricePoint, error) {
	var history []shared.PricePoint
	if i.queryAPI == nil {
		return history, nil
	}
	result, err := i.queryAPI.Query(ctx, priceHistoryQuery(i.bucket, id))
	if err != nil {
		return history, fmt.Errorf("failed to query price history: %w", err)
	}
	defer result.Close()
	for result.Next() {
		record := result.Record()
		var cents int
		switch v := record.Value().(type) {
		case int64:
			cents = int(v)
		case float64:
			cents = int(v)
		default:
			continue
		}
		history = append(history, shared.PricePoint{Time: record.Time(), PriceCents: cents})
	}
	if result.Err() != nil {
		return history, fmt.Errorf("failed to read price history: %w", result.Err())
	}
	return history, nil
}

func (i *InfluxDB) WriteProductDatapoint(info shared.ProductInfo) {
	values := map[string]interface{}{"cents": info.PriceCents, "grams": info.WeightGrams}

	// If the price has meaningfully changed, report the change
	if info.PriceCents != 0 && info.PreviousPriceCents != 0 && info.PriceCents != info.PreviousPriceCents {
		values["cents_change"] = info.PriceCents - info.PreviousPriceCents
	}
	p := influxdb2.NewPoint("product",
		map[string]string{
			"name":       info.Name,
			"store":      info.Store,
			"location":   info.Location,
			"department": info.Department,
			"id":         info.ID,
		},
		values,
		info.Timestamp,
	)
	i.groceryWriteAPI.WritePoint(p)
}

func (i *InfluxDB) WriteArbitrarySystemDatapoint(field string, value interface{}) {
	p := influxdb2.NewPoint("system",
		map[string]string{"service": shared.SYSTEM_SERVICE_NAME},
		map[string]interface{}{field: value},
		time.Now(),
	)
	i.systemWriteAPI.WritePoint(p)
}

func (i *InfluxDB) WriteSystemDatapoint(data shared.SystemStatusDatapoint) {
	p := influxdb2.NewPoint("system",
		map[string]string{},
		map[string]interface{}{
			shared.SYSTEM_RAM_UTILISATION_PERCENT_FIELD: data.RAMUtilisationPercent,
			shared.SYSTEM_PRODUCTS_PER_SECOND_FIELD:     data.ProductsPerSecond,
			shared.SYSTEM_HDD_BYTES_FREE_FIELD:          data.HDDBytesFree,
			shared.SYSTEM_TOTAL_PRODUCT_COUNT_FIELD:     data.TotalProductCount,
		},
		time.Now(),
	)
	i.systemWriteAPI.WritePoint(p)
}

// WriteWorker writes ProductInfo to InfluxDB
// Note that the underlying library automatically batches writes
// so we don't need to worry about that here.
func (i *InfluxDB) WriteWorker(input <-chan shared.ProductInfo) {
	for info := range input {
		i.WriteProductDatapoint(info)
	}
}

func (i *InfluxDB) Close() {
	i.groceryWriteAPI.Flush()
	i.systemWriteAPI.Flush()
	i.db.Close()
	if i.queryDB != nil {
		i.queryDB.Close()
	}
}
