package influxdb

import (
	"log/slog"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	shared "github.com/tjhowse/aus_grocery_price_database/internal/shared"
)

type InfluxDB struct {
	db              influxdb2.Client
	groceryWriteAPI api.WriteAPI
	systemWriteAPI  api.WriteAPI
}

func (i *InfluxDB) Init(url, token, org, bucket string) {
	slog.Info("Initialising InfluxDB", "url", url, "org", org, "bucket", bucket)
	i.db = influxdb2.NewClient(url, token)
	i.groceryWriteAPI = i.db.WriteAPI(org, bucket)
	i.systemWriteAPI = i.db.WriteAPI(org, "system")
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
}
