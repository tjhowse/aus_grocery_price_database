package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	woolworths "github.com/tjhowse/aus_grocery_price_database/internal/woolworths"
)

type WoolworthsWorker struct {
	connection woolworths.Woolworths
	output     chan woolworths.ProductInfo
}

func (w *WoolworthsWorker) Init() chan woolworths.ProductInfo {
	w.connection = woolworths.Woolworths{}
	w.connection.Init("https://www.woolworths.com.au")
	w.output = make(chan woolworths.ProductInfo)
	return w.output
}
func (w *WoolworthsWorker) Run(input chan woolworths.ProductID) {
	slog.Debug("Running WoolworthsWorker")
	for id := range input {
		slog.Debug(fmt.Sprintf("Getting product info for ID: %d", id))
		info, err := w.connection.GetProductInfo(id)
		if err != nil {
			slog.Error(fmt.Sprintf("Error getting product info: %v", err))
		}
		w.output <- info
	}
}

// Convert from woolworths.ProductInfo to main.ProductInfo
func ConvertWoolworthsProductInfo(wProductInfo woolworths.ProductInfo) ProductInfo {
	return ProductInfo{
		Description: wProductInfo.Description,
		Price:       wProductInfo.Offers.Price,
		WeightGrams: wProductInfo.Weight,
	}
}

func StartWorker() {
	// Create a new WoolworthsWorker
	worker := WoolworthsWorker{}
	// Create a channel to send ProductIDs to the worker
	input := make(chan woolworths.ProductID)
	// Initialize the worker
	output := worker.Init()
	// Start the worker
	go worker.Run(input)

	// Send a ProductID to the worker
	slog.Debug("Sending ProductID to worker")
	input <- woolworths.ProductID(133211)
	slog.Debug("Sent ProductID to worker")

	slog.Info(fmt.Sprintf("Product Info: %v", <-output))
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
	w.Init("https://www.woolworths.com.au")

	if prodids, err := w.GetProductList(); err != nil {
		slog.Error(fmt.Sprintf("Error getting product list: %v", err))
	} else {
		slog.Info(fmt.Sprintf("Product IDs: %d", prodids[0]))
	}
	// StartWorker()

}
