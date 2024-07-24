package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	shared "github.com/tjhowse/aus_grocery_price_database/internal/shared"
	woolworths "github.com/tjhowse/aus_grocery_price_database/internal/woolworths"
)

const VERSION = "0.0.3"

type config struct {
	InfluxDBURL           string `env:"INFLUXDB_URL"`
	InfluxDBToken         string `env:"INFLUXDB_TOKEN"`
	InfluxDBOrg           string `env:"INFLUXDB_ORG" envDefault:"groceries"`
	InfluxDBBucket        string `env:"INFLUXDB_BUCKET" envDefault:"groceries"`
	LocalWoolworthsDBPath string `env:"LOCAL_WOOLWORTHS_DB_PATH" envDefault:"woolworths.db3"`
	MaxProductAgeMinutes  int    `env:"MAX_PRODUCT_AGE_MINUTES" envDefault:"1440"`
	WoolworthsURL         string `env:"WOOLWORTHS_URL" envDefault:"https://www.woolworths.com.au"`
	DebugLogging          string `env:"DEBUG_LOGGING" envDefault:"false"`
}

const INFLUX_UPDATE_RATE_SECONDS = 10

// Other stores will need to implement this interface
type ProductInfoGetter interface {
	Init(string, string, time.Duration)
	Run(chan struct{})
	GetSharedProductsUpdatedAfter(time.Time, int) ([]shared.ProductInfo, error)
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

	influx.WriteArbitrarySystemDatapoint(SYSTEM_VERSION_FIELD, VERSION)

	products := make(chan shared.ProductInfo)
	go influx.WriteWorker(products)
	defer close(products)

	cancel := make(chan struct{})
	defer close(cancel)
	go w.Run(cancel)

	var systemStatus SystemStatusDatapoint

	// Assume we were shut down for half an hour.
	// TODO Store the last update time in a main-level database.
	updateTime := time.Now().Add(-30 * time.Minute)
	for {
		woolworthsProducts, err := w.GetSharedProductsUpdatedAfter(updateTime, 100)
		if err != nil {
			slog.Error("Error getting shared products", "error", err)
			time.Sleep(10 * time.Second)
			continue
		}
		if len(woolworthsProducts) != 0 {
			updateTime = time.Now()
		}
		for _, product := range woolworthsProducts {
			slog.Info("Updating product data", "name", product.Name, "price", product.PriceCents)
			products <- product
		}
		systemStatus.ProductsPerSecond = float64(len(woolworthsProducts)) / INFLUX_UPDATE_RATE_SECONDS
		systemStatus.RAMUtilisationPercent = GetRAMUtilisationPercent()
		influx.WriteSystemDatapoint(systemStatus)
		time.Sleep(INFLUX_UPDATE_RATE_SECONDS * time.Second)
	}
}
