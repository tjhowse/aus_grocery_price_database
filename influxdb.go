package main

import (
	"log/slog"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	shared "github.com/tjhowse/aus_grocery_price_database/internal/shared"
)

const SYSTEM_VERSION_FIELD = "version"
const SYSTEM_SERVICE_NAME = "agpd"
const SYSTEM_RAM_UTILISATION_PERCENT_FIELD = "ram_utilisation_percentage"
const SYSTEM_PRODUCTS_PER_SECOND_FIELD = "products_per_second"
const SYSTEM_HDD_BYTES_FREE_FIELD = "hdd_bytes_free"
const SYSTEM_TOTAL_PRODUCT_COUNT_FIELD = "total_product_count"

type SystemStatusDatapoint struct {
	RAMUtilisationPercent float64
	ProductsPerSecond     float64
	HDDBytesFree          int
	TotalProductCount     int
}

type influxDB struct {
	db              influxdb2.Client
	groceryWriteAPI api.WriteAPI
	systemWriteAPI  api.WriteAPI
}

func (i *influxDB) Init(url, token, org, bucket string) {
	slog.Info("Initialising InfluxDB", "url", url, "org", org, "bucket", bucket)
	i.db = influxdb2.NewClient(url, token)
	i.groceryWriteAPI = i.db.WriteAPI(org, bucket)
	i.systemWriteAPI = i.db.WriteAPI(org, "system")
}

func (i *influxDB) WriteProductDatapoint(info shared.ProductInfo) {
	p := influxdb2.NewPoint("product",
		map[string]string{"name": info.Name, "store": info.Store, "location": info.Location, "department": info.Department},
		map[string]interface{}{"cents": info.PriceCents, "grams": info.WeightGrams},
		info.Timestamp,
	)
	i.groceryWriteAPI.WritePoint(p)
}

func (i *influxDB) WriteArbitrarySystemDatapoint(field string, value interface{}) {
	p := influxdb2.NewPoint("system",
		map[string]string{"service": SYSTEM_SERVICE_NAME},
		map[string]interface{}{field: value},
		time.Now(),
	)
	i.systemWriteAPI.WritePoint(p)
}

func (i *influxDB) WriteSystemDatapoint(data SystemStatusDatapoint) {
	p := influxdb2.NewPoint("system",
		map[string]string{},
		map[string]interface{}{
			SYSTEM_RAM_UTILISATION_PERCENT_FIELD: data.RAMUtilisationPercent,
			SYSTEM_PRODUCTS_PER_SECOND_FIELD:     data.ProductsPerSecond,
			SYSTEM_HDD_BYTES_FREE_FIELD:          data.HDDBytesFree,
			SYSTEM_TOTAL_PRODUCT_COUNT_FIELD:     data.TotalProductCount,
		},
		time.Now(),
	)
	i.systemWriteAPI.WritePoint(p)
}

// WriteWorker writes ProductInfo to InfluxDB
// Note that the underlying library automatically batches writes
// so we don't need to worry about that here.
func (i *influxDB) WriteWorker(input <-chan shared.ProductInfo) {
	for info := range input {
		i.WriteProductDatapoint(info)
	}
}

func (i *influxDB) Close() {
	i.groceryWriteAPI.Flush()
	i.systemWriteAPI.Flush()
	i.db.Close()
}
