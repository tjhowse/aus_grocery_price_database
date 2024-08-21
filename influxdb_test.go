package main

import (
	"testing"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	shared "github.com/tjhowse/aus_grocery_price_database/internal/shared"
)

type MockInfluxdbWriteAPI struct {
	writtenPoints []*write.Point
}

func (m *MockInfluxdbWriteAPI) WritePoint(p *write.Point) {
	m.writtenPoints = append(m.writtenPoints, p)
}

func (m *MockInfluxdbWriteAPI) WriteRecord(line string) {}

func (m *MockInfluxdbWriteAPI) Flush() {}

func (m *MockInfluxdbWriteAPI) Errors() <-chan error {
	return make(chan error)
}

// func (m *MockInfluxdbWriteAPI) Init() {
// 	m.writtenPoints = []*write.Point{}
// }

func (m *MockInfluxdbWriteAPI) SetWriteFailedCallback(cb api.WriteFailedCallback) {}

func InitMockInfluxDB() (*influxDB, *MockInfluxdbWriteAPI, *MockInfluxdbWriteAPI) {
	i := influxDB{}
	groceryMock := &MockInfluxdbWriteAPI{}
	i.groceryWriteAPI = groceryMock
	systemMock := &MockInfluxdbWriteAPI{}
	i.systemWriteAPI = systemMock

	return &i, groceryMock, systemMock
}

func TestWriteProductDatapoint(t *testing.T) {
	desiredTags := map[string]string{
		"name":       "Test Product",
		"store":      "Test Store",
		"location":   "Test Location",
		"department": "Test Department",
	}
	i, gMock, _ := InitMockInfluxDB()
	i.WriteProductDatapoint(shared.ProductInfo{
		Name:               desiredTags["name"],
		Store:              desiredTags["store"],
		Location:           desiredTags["location"],
		Department:         desiredTags["department"],
		PriceCents:         100,
		PreviousPriceCents: 0,
		WeightGrams:        1000,
		Timestamp:          time.Now(),
	})
	i.WriteProductDatapoint(shared.ProductInfo{
		Name:               desiredTags["name"],
		Store:              desiredTags["store"],
		Location:           desiredTags["location"],
		Department:         desiredTags["department"],
		PriceCents:         101,
		PreviousPriceCents: 100,
		WeightGrams:        1000,
		Timestamp:          time.Now(),
	})
	i.WriteProductDatapoint(shared.ProductInfo{
		Name:               desiredTags["name"],
		Store:              desiredTags["store"],
		Location:           desiredTags["location"],
		Department:         desiredTags["department"],
		PriceCents:         99,
		PreviousPriceCents: 101,
		WeightGrams:        1000,
		Timestamp:          time.Now(),
	})

	if want, got := 3, len(gMock.writtenPoints); want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	// Check the first written point.
	p := gMock.writtenPoints[0]
	if want, got := "product", p.Name(); want != got {
		t.Errorf("want %s, got %s", want, got)
	}

	for _, tag := range p.TagList() {
		if want, got := desiredTags[tag.Key], tag.Value; want != got {
			t.Errorf("want %s, got %s", want, got)
		}
	}

	for _, field := range p.FieldList() {
		switch field.Key {
		case "cents":
			if want, got := int64(100), field.Value.(int64); want != got {
				t.Errorf("want %v, got %v", want, got)
			}
		case "grams":
			if want, got := int64(1000), field.Value.(int64); want != got {
				t.Errorf("want %v, got %v", want, got)
			}
		case "cents_change":
			t.Errorf("unexpected field %s", field.Key)
		default:
			t.Errorf("unexpected field %s", field.Key)
		}
	}

	// Now check the second written point.
	p = gMock.writtenPoints[1]

	for _, field := range p.FieldList() {
		switch field.Key {
		case "cents":
			if want, got := int64(101), field.Value.(int64); want != got {
				t.Errorf("want %v, got %v", want, got)
			}
		case "grams":
			continue
		case "cents_change":
			if want, got := int64(1), field.Value.(int64); want != got {
				t.Errorf("want %v, got %v", want, got)
			}
		default:
			t.Errorf("unexpected field %s", field.Key)
		}
	}

	// Now check the third written point.
	p = gMock.writtenPoints[2]

	for _, field := range p.FieldList() {
		switch field.Key {
		case "cents":
			continue
		case "grams":
			continue
		case "cents_change":
			if want, got := int64(-2), field.Value.(int64); want != got {
				t.Errorf("want %v, got %v", want, got)
			}
		default:
			t.Errorf("unexpected field %s", field.Key)
		}
	}

}
