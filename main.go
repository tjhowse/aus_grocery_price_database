package main

import (
	"flag"
	"log/slog"
	"os"
	"time"

	woolworths "github.com/tjhowse/aus_grocery_price_database/internal/woolworths"
)

// Convert from woolworths.ProductInfo to main.ProductInfo
func ConvertWoolworthsProductInfo(wProductInfo woolworths.ProductInfo) ProductInfo {
	return ProductInfo{
		Description: wProductInfo.Description,
		Price:       wProductInfo.Offers.Price,
		WeightGrams: wProductInfo.Weight,
	}
}
func main() {
	// var err error
	verbose := flag.Bool("v", false, "verbose")
	flag.Parse()
	logLevel := slog.LevelInfo
	if *verbose {
		// Set the log level to debug
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	slog.Debug("Hi!")

	w := woolworths.Woolworths{}
	// w.Init("https://www.woolworths.com.au", ":memory:", woolworths.PRODUCT_INFO_MAX_AGE)
	// w.Init("https://www.woolworths.com.au", "woolworths.db3", woolworths.PRODUCT_INFO_MAX_AGE)
	w.Init("https://www.woolworths.com.au", "woolworths.db3", 5*time.Minute)
	cancel := make(chan struct{})
	go w.RunScheduler(cancel)
	time.Sleep(60 * time.Minute)
	close(cancel)
}
