package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	woolworths "github.com/tjhowse/aus_grocery_price_database/internal/woolworths"
)

// TODO https://github.com/influxdata/influxdb-client-go
type config struct {
	InfluxDBURL           string `env:"INFLUXDB_URL"`
	InfluxDBToken         string `env:"INFLUXDB_TOKEN"`
	InfluxDBOrg           string `env:"INFLUXDB_ORG"`
	InfluxDBBucker        string `env:"INFLUXDB_BUCKER"`
	LocalWoolworthsDBPath string `env:"LOCAL_WOOLWORTHS_DB_PATH" envDefault:":memory:"`
	MaxProductAgeMinutes  int    `env:"MAX_PRODUCT_AGE_MINUTES" envDefault:"1440"`
	WoolworthsURL         string `env:"WOOLWORTHS_URL" envDefault:"https://www.woolworths.com.au"`
}

// Convert from woolworths.ProductInfo to main.ProductInfo
func ConvertWoolworthsProductInfo(wProductInfo woolworths.ProductInfo) ProductInfo {
	return ProductInfo{
		Description: wProductInfo.Description,
		Price:       wProductInfo.Offers.Price,
		WeightGrams: wProductInfo.Weight,
	}
}
func main() {

	// Read in the environment variables
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}
	verbose := flag.Bool("v", false, "verbose")
	flag.Parse()
	logLevel := slog.LevelInfo
	if *verbose || cfg.DebugLogging == "true" {
		// Set the log level to debug
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	slog.Info("AUS Grocery Price Database", "version", VERSION)

	w := woolworths.Woolworths{}
	w.Init(cfg.WoolworthsURL, cfg.LocalWoolworthsDBPath, time.Duration(cfg.MaxProductAgeMinutes)*time.Minute)

	influx := influxDB{}
	influx.Init(cfg.InfluxDBURL, cfg.InfluxDBToken, cfg.InfluxDBOrg, cfg.InfluxDBBucket)
	defer influx.Close()

	products := make(chan shared.ProductInfo)
	go influx.WriteWorker(products)
	defer close(products)

	cancel := make(chan struct{})
	defer close(cancel)
	go w.RunScheduler(cancel)

	for {
		updateTime := time.Now()
		_, err := w.GetProductIDsUpdatedAfter(updateTime, 10)
		if err != nil {
			slog.Error("Error getting product IDs", "error", err)
			time.Sleep(10 * time.Second)
			continue
		}
		// for _, productID := range woolworthsProducts {

	}

}
