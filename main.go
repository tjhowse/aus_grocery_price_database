package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

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
	w.Init("https://www.woolworths.com.au", ":memory:")

	// if prodids, err := w.GetProductList(); err != nil {
	// 	slog.Error(fmt.Sprintf("Error getting product list: %v", err))
	// } else {
	// 	slog.Info(fmt.Sprintf("Product IDs: %d", prodids[0]))
	// }
	// StartWorker()

	inputChannel := make(chan woolworths.ProductID)
	outputChannel := make(chan woolworths.WoolworthsProductInfo)
	go w.ProductWorker(inputChannel, outputChannel)
	go w.ProductWorker(inputChannel, outputChannel)
	inputChannel <- woolworths.ProductID(187314)
	inputChannel <- woolworths.ProductID(187315)
	slog.Info(fmt.Sprintf("Product Info: %v", <-outputChannel))
	slog.Info(fmt.Sprintf("Product Info: %v", <-outputChannel))

}
