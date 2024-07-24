package shared

import "time"

type ProductInfo struct {
	ID          string
	Name        string
	Description string
	Store       string
	Department  string
	Location    string
	PriceCents  int
	WeightGrams int
	Timestamp   time.Time
}
