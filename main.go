package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/tjhowse/aus_grocery_price_database/internal/coles"
	shared "github.com/tjhowse/aus_grocery_price_database/internal/shared"
	woolworths "github.com/tjhowse/aus_grocery_price_database/internal/woolworths"
)

const VERSION = "0.0.40"
const SYSTEM_STATUS_UPDATE_INTERVAL_SECONDS = 60

type config struct {
	InfluxDBURL                 string `env:"INFLUXDB_URL"`
	InfluxDBToken               string `env:"INFLUXDB_TOKEN"`
	InfluxDBOrg                 string `env:"INFLUXDB_ORG" envDefault:"groceries"`
	InfluxDBBucket              string `env:"INFLUXDB_BUCKET" envDefault:"groceries"`
	InfluxUpdateIntervalSeconds int    `env:"INFLUXDB_UPDATE_RATE_SECONDS" envDefault:"10"`
	LocalWoolworthsDBPath       string `env:"LOCAL_WOOLWORTHS_DB_PATH" envDefault:"woolworths.db3"`
	LocalColesDBPath            string `env:"LOCAL_COLES_DB_PATH" envDefault:"coles.db3"`
	MaxProductAgeMinutes        int    `env:"MAX_PRODUCT_AGE_MINUTES" envDefault:"1440"`
	WoolworthsURL               string `env:"WOOLWORTHS_URL" envDefault:"https://www.woolworths.com.au"`
	ColesURL                    string `env:"COLES_URL" envDefault:"https://www.coles.com.au"`
	DebugLogging                bool   `env:"DEBUG_LOGGING" envDefault:"false"`
}

// ProductInfoGetter defines the expectations for a product information getter.
type ProductInfoGetter interface {
	Init(string, string, time.Duration) error
	Run(chan struct{})
	GetSharedProductsUpdatedAfter(time.Time, int) ([]shared.ProductInfo, error)
	GetTotalProductCount() (int, error)
}

type timeseriesDB interface {
	Init(string, string, string, string)
	WriteProductDatapoint(shared.ProductInfo)
	WriteArbitrarySystemDatapoint(string, interface{})
	WriteSystemDatapoint(SystemStatusDatapoint)
	WriteWorker(<-chan shared.ProductInfo)
	Close()
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
	if *verbose || cfg.DebugLogging {
		// Set the log level to debug
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	slog.Info("AUS Grocery Price Database", "version", VERSION)

	tsDB := influxDB{}
	tsDB.Init(cfg.InfluxDBURL, cfg.InfluxDBToken, cfg.InfluxDBOrg, cfg.InfluxDBBucket)
	defer tsDB.Close()

	w := woolworths.Woolworths{}
	w.Init(cfg.WoolworthsURL, cfg.LocalWoolworthsDBPath, time.Duration(cfg.MaxProductAgeMinutes)*time.Minute)

	c := coles.Coles{}
	c.Init(cfg.ColesURL, cfg.LocalColesDBPath, time.Duration(cfg.MaxProductAgeMinutes)*time.Minute)

	running := true
	run(&running, &cfg, &tsDB, []ProductInfoGetter{&w, &c})

}

func run(running *bool, cfg *config, tsDB timeseriesDB, pigs []ProductInfoGetter) {
	var err error

	tsDB.WriteArbitrarySystemDatapoint(SYSTEM_VERSION_FIELD, VERSION)

	productInfoUpdateChannel := make(chan shared.ProductInfo)
	go tsDB.WriteWorker(productInfoUpdateChannel)
	defer close(productInfoUpdateChannel)

	cancel := make(chan struct{})
	defer close(cancel)
	for _, pig := range pigs {
		pig.Run(cancel)
	}

	updateTime := time.Now().Add(-1 * time.Minute)
	var updateCountSinceLastStatusReport int

	var systemStatus SystemStatusDatapoint
	// Ensure a status update is sent out immediately.
	statusReportDeadline := time.Now().Add(-30 * time.Minute)

	for *running {
		// Get the latest products from the grocery stores.
		products := make([]shared.ProductInfo, 0, 200)
		for _, pig := range pigs {
			prods, err := pig.GetSharedProductsUpdatedAfter(updateTime, 100)
			if err != nil {
				slog.Error("Error getting shared products", "error", err)
				time.Sleep(10 * time.Second)
				continue
			}
			products = append(products, prods...)
		}
		if len(products) != 0 {
			updateTime = time.Now()
		}
		for _, newProductInfo := range products {
			if newProductInfo.Name == "" {
				slog.Warn("Product has no name", "product", newProductInfo)
				continue
			}
			productInfoUpdateChannel <- newProductInfo
		}

		updateCountSinceLastStatusReport += len(products)

		// Send a system status update if required.
		if time.Now().After(statusReportDeadline) {
			systemStatus.ProductsPerSecond = float64(updateCountSinceLastStatusReport) / SYSTEM_STATUS_UPDATE_INTERVAL_SECONDS
			updateCountSinceLastStatusReport = 0

			systemStatus.RAMUtilisationPercent = GetRAMUtilisationPercent()
			systemStatus.HDDBytesFree, err = GetHDDBytesFree()
			if err != nil {
				slog.Error("Error getting HDD free space", "error", err)
			}
			// Total up all the products in the system.
			systemStatus.TotalProductCount = 0
			for _, pig := range pigs {
				count, err := pig.GetTotalProductCount()
				if err != nil {
					slog.Error("Error getting total product count", "error", err)
				}
				systemStatus.TotalProductCount += count
			}
			tsDB.WriteSystemDatapoint(systemStatus)
			statusReportDeadline = time.Now().Add(SYSTEM_STATUS_UPDATE_INTERVAL_SECONDS * time.Second)
			slog.Info("Heartbeat", "productsPerSecond", systemStatus.ProductsPerSecond)
		}
		time.Sleep(time.Duration(cfg.InfluxUpdateIntervalSeconds) * time.Second)
	}
}
