package shared

import (
	"errors"
	"time"
)

var ErrProductMissing = errors.New("no product found")

// ProductInfo is a struct that contains information about a product.
type ProductInfo struct {
	ID                 string
	Name               string
	Description        string
	Store              string
	Department         string
	Location           string
	PriceCents         int
	PreviousPriceCents int
	WeightGrams        int
	Timestamp          time.Time
}
