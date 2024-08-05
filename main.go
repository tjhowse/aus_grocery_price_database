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

const VERSION = "0.0.27"
const SYSTEM_STATUS_UPDATE_INTERVAL_SECONDS = 60

type config struct {
	InfluxDBURL                 string `env:"INFLUXDB_URL"`
	InfluxDBToken               string `env:"INFLUXDB_TOKEN"`
	InfluxDBOrg                 string `env:"INFLUXDB_ORG" envDefault:"groceries"`
	InfluxDBBucket              string `env:"INFLUXDB_BUCKET" envDefault:"groceries"`
	InfluxUpdateIntervalSeconds int    `env:"INFLUXDB_UPDATE_RATE_SECONDS" envDefault:"10"`
	LocalWoolworthsDBPath       string `env:"LOCAL_WOOLWORTHS_DB_PATH" envDefault:"woolworths.db3"`
	MaxProductAgeMinutes        int    `env:"MAX_PRODUCT_AGE_MINUTES" envDefault:"1440"`
	WoolworthsURL               string `env:"WOOLWORTHS_URL" envDefault:"https://www.woolworths.com.au"`
	DebugLogging                bool   `env:"DEBUG_LOGGING" envDefault:"false"`
}

// Other stores will need to implement this interface
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
	w := woolworths.Woolworths{}
	running := true
	run(&running, &cfg, &tsDB, &w)
}

func run(running *bool, cfg *config, tsDB timeseriesDB, w ProductInfoGetter) {
	w.Init(cfg.WoolworthsURL, cfg.LocalWoolworthsDBPath, time.Duration(cfg.MaxProductAgeMinutes)*time.Minute)

	tsDB.Init(cfg.InfluxDBURL, cfg.InfluxDBToken, cfg.InfluxDBOrg, cfg.InfluxDBBucket)
	defer tsDB.Close()

	tsDB.WriteArbitrarySystemDatapoint(SYSTEM_VERSION_FIELD, VERSION)

	productInfoUpdateChannel := make(chan shared.ProductInfo)
	go tsDB.WriteWorker(productInfoUpdateChannel)
	defer close(productInfoUpdateChannel)

	cancel := make(chan struct{})
	defer close(cancel)
	go w.Run(cancel)

	updateTime := time.Now().Add(-1 * time.Minute)
	var updateCountSinceLastStatusReport int

	var systemStatus SystemStatusDatapoint
	// Ensure a status update is sent out immediately.
	statusReportDeadline := time.Now().Add(-30 * time.Minute)

	for *running {
		woolworthsProducts, err := w.GetSharedProductsUpdatedAfter(updateTime, 100)
		if err != nil {
			slog.Error("Error getting shared products", "error", err)
			time.Sleep(10 * time.Second)
			continue
		}
		if len(woolworthsProducts) != 0 {
			updateTime = time.Now()
		}
		for _, newProductInfo := range woolworthsProducts {
			if newProductInfo.Name == "" {
				slog.Warn("Product has no name", "product", newProductInfo)
				continue
			}
			slog.Info("Updating product data", "name", newProductInfo.Name, "price", newProductInfo.PriceCents)
			productInfoUpdateChannel <- newProductInfo
		}

		updateCountSinceLastStatusReport += len(woolworthsProducts)

		// Send a system status update if required.
		if time.Now().After(statusReportDeadline) {
			systemStatus.ProductsPerSecond = float64(updateCountSinceLastStatusReport) / SYSTEM_STATUS_UPDATE_INTERVAL_SECONDS
			updateCountSinceLastStatusReport = 0

			systemStatus.RAMUtilisationPercent = GetRAMUtilisationPercent()
			systemStatus.HDDBytesFree, err = GetHDDBytesFree()
			if err != nil {
				slog.Error("Error getting HDD free space", "error", err)
			}

			systemStatus.TotalProductCount, err = w.GetTotalProductCount()
			if err != nil {
				slog.Error("Error getting total product count", "error", err)
			}
			tsDB.WriteSystemDatapoint(systemStatus)
			statusReportDeadline = time.Now().Add(SYSTEM_STATUS_UPDATE_INTERVAL_SECONDS * time.Second)
			slog.Debug("Heartbeat")
		}
		time.Sleep(time.Duration(cfg.InfluxUpdateIntervalSeconds) * time.Second)
	}
}
