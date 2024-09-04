package coles

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestCalcWeightInGrams(t *testing.T) {
	c := getInitialisedColes()
	products, _, err := c.getProductsAndTotalCountForCategoryPage("fruit-vegetables", 1)
	if err != nil {
		t.Fatalf("Failed to get products: %v", err)
	}

	var cases = []struct {
		index int
		want  int
		err   bool
	}{
		{0, 1000, false},
		{7, 170, false},
		{8, 0, true},
	}

	for _, tc := range cases {
		weight, err := calcWeightInGrams(products[tc.index])
		if err != nil != tc.err {
			t.Fatalf("Unexpectedly failed to calculate weight for item %s: %v", products[tc.index].Info.Name, err)
		}
		if want, got := tc.want, weight; want != got {
			t.Errorf("Expected %d, got %d  for test item %s", want, got, products[tc.index].Info.Name)
		}
	}
}

func TestSaveProductInfo(t *testing.T) {
	c := getInitialisedColes()
	products, _, err := c.getProductsAndTotalCountForCategoryPage("fruit-vegetables", 1)
	if err != nil {
		t.Fatalf("Failed to get products: %v", err)
	}

	tx, err := c.db.Begin()
	if err != nil {
		t.Fatalf("Failed to start transaction: %v", err)
	}

	savedProduct := products[0]
	c.saveProductInfo(tx, savedProduct)

	if err = tx.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Check if the loadedProduct was saved
	loadedProduct, err := c.loadProductInfo(savedProduct.ID)
	if err != nil {
		t.Fatalf("Failed to load product info: %v", err)
	}

	// Check name
	if want, got := savedProduct.Info.Name, loadedProduct.Info.Name; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
	// Check description
	if want, got := savedProduct.Info.Description, loadedProduct.Info.Description; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
	// Check price
	if want, got := savedProduct.Info.Pricing.Now.Mul(decimal.NewFromInt(100)).IntPart(), loadedProduct.Info.Pricing.Now.IntPart(); want != got {
		t.Errorf("Expected %d, got %d", want, got)
	}
	savedProductWeight, err := calcWeightInGrams(savedProduct)
	if err != nil {
		t.Fatalf("Failed to calculate weight: %v", err)
	}
	// Check weight
	if want, got := savedProductWeight, loadedProduct.WeightGrams; want != got {
		t.Errorf("Expected %d, got %d", want, got)
	}
	// Check departmentID
	if want, got := "fruit-vegetables", loadedProduct.departmentID; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}

}

func TestDepartmentInfo(t *testing.T) {
	w := getInitialisedColes()
	dept := departmentInfo{SeoToken: "1-E5BEE36E", Name: "Fruit & Veg", Updated: time.Now()}
	w.saveDepartment(dept)
	departmentIDs, err := w.loadDepartmentInfoList()
	if err != nil {
		t.Fatal(err)
	}
	if want, got := 1, len(departmentIDs); want != got {
		t.Errorf("Expected %d, got %d", want, got)
	}
	if want, got := "1-E5BEE36E", departmentIDs[0].SeoToken; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
	if want, got := "Fruit & Veg", departmentIDs[0].Name; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
}
