package main

import (
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	shared "github.com/tjhowse/aus_grocery_price_database/internal/shared"
)

type influxDB struct {
	db       influxdb2.Client
	writeAPI api.WriteAPI
}

func (i *influxDB) Init(url, token, org, bucket string) {
	i.db = influxdb2.NewClient(url, token)
	i.writeAPI = i.db.WriteAPI(org, bucket)
}

func (i *influxDB) WriteDatapoint(info shared.ProductInfo) {
	p := influxdb2.NewPoint("product",
		map[string]string{"name": info.Name, "store": info.Store, "location": info.Location},
		map[string]interface{}{"price": info.PriceCents, "weight": info.WeightGrams},
		info.Timestamp,
	)
	i.writeAPI.WritePoint(p)
}

// WriteWorker writes ProductInfo to InfluxDB
// Note that the underlying library automatically batches writes
// so we don't need to worry about that here.
func (i *influxDB) WriteWorker(input <-chan shared.ProductInfo) {
	for info := range input {
		i.WriteDatapoint(info)
	}
}

func (i *influxDB) Close() {
	i.writeAPI.Flush()
	i.db.Close()
}
