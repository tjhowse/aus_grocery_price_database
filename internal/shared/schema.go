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

const SYSTEM_VERSION_FIELD = "version"
const SYSTEM_SERVICE_NAME = "agpd"
const SYSTEM_RAM_UTILISATION_PERCENT_FIELD = "ram_utilisation_percentage"
const SYSTEM_PRODUCTS_PER_SECOND_FIELD = "products_per_second"
const SYSTEM_HDD_BYTES_FREE_FIELD = "hdd_bytes_free"
const SYSTEM_TOTAL_PRODUCT_COUNT_FIELD = "total_product_count"

type SystemStatusDatapoint struct {
	RAMUtilisationPercent float64
	ProductsPerSecond     float64
	HDDBytesFree          int
	TotalProductCount     int
}
