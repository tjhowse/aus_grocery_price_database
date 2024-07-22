package shared

import "time"

type ProductInfo struct {
	Name        string
	Description string
	Store       string
	Location    string
	Price       float32
	WeightGrams float32
	Timestamp   time.Time
}
