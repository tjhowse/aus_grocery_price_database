package coles

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestCalcWeightInGrams(t *testing.T) {
	c := getInitialisedColes()
	dp := departmentPage{"fruit-vegetables", 1}
	products, _, err := c.getProductsAndTotalCountForCategoryPage(dp)
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
	dp := departmentPage{"fruit-vegetables", 1}
	products, _, err := c.getProductsAndTotalCountForCategoryPage(dp)
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

func TestGetSharedProductsUpdatedAfter(t *testing.T) {
	c := Coles{}
	c.Init(colesServer.URL, ":memory:", 5*time.Second)
	c.filterDepartments = false
	var infoList []colesProductInfo
	infoList = append(infoList, colesProductInfo{ID: "123455", Info: productListPageProduct{Name: "1", Pricing: productListPageProductPricing{Now: decimal.NewFromFloat(1.5)}}, Updated: time.Now().Add(-5 * time.Minute)})
	infoList = append(infoList, colesProductInfo{ID: "123456", Info: productListPageProduct{Name: "2", Pricing: productListPageProductPricing{Now: decimal.NewFromFloat(2.4)}}, Updated: time.Now().Add(-4 * time.Minute)})
	infoList = append(infoList, colesProductInfo{ID: "123457", Info: productListPageProduct{Name: "3", Pricing: productListPageProductPricing{Now: decimal.NewFromFloat(3.3)}}, Updated: time.Now().Add(-3 * time.Minute)})
	infoList = append(infoList, colesProductInfo{ID: "123458", Info: productListPageProduct{Name: "4", Pricing: productListPageProductPricing{Now: decimal.NewFromFloat(4.2)}}, Updated: time.Now().Add(-1 * time.Minute)})
	// Put this one in twice to test the PreviousPriceCents is updated.
	infoList = append(infoList, colesProductInfo{ID: "123459", Info: productListPageProduct{Name: "5", Pricing: productListPageProductPricing{Now: decimal.NewFromFloat(5.0)}}, Updated: time.Now()})
	infoList = append(infoList, colesProductInfo{ID: "123459", Info: productListPageProduct{Name: "5", Pricing: productListPageProductPricing{Now: decimal.NewFromFloat(5.1)}}, Updated: time.Now()})
	// This last one is to test that we don't get products that have a blank name.
	infoList = append(infoList, colesProductInfo{ID: "123460", Info: productListPageProduct{Name: "", Pricing: productListPageProductPricing{Now: decimal.NewFromFloat(6.0)}}, Updated: time.Now()})

	c.saveProductInfoes(infoList)

	productIDs, err := c.GetSharedProductsUpdatedAfter(time.Now().Add(-2*time.Minute), 10)
	if err != nil {
		t.Fatal(err)
	}
	if want, got := 2, len(productIDs); want != got {
		t.Fatalf("Expected %d products, got %d", want, got)
	}
	if want, got := COLES_ID_PREFIX+"123458", productIDs[0].ID; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
	if want, got := COLES_ID_PREFIX+"123459", productIDs[1].ID; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
	if want, got := 500, productIDs[1].PreviousPriceCents; want != got {
		t.Errorf("Expected %v, got %v", want, got)
	}
	if want, got := 510, productIDs[1].PriceCents; want != got {
		t.Errorf("Expected %v, got %v", want, got)
	}
	productIDs, err = c.GetSharedProductsUpdatedAfter(time.Now().Add(-2*time.Minute), 1)
	if err != nil {
		t.Fatal(err)
	}
	if want, got := 1, len(productIDs); want != got {
		t.Fatalf("Expected %d products, got %d", want, got)
	}
	if want, got := COLES_ID_PREFIX+"123458", productIDs[0].ID; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
	if want, got := 420, productIDs[0].PriceCents; want != got {
		t.Errorf("Expected %v, got %v", want, got)
	}

	if total, err := c.GetTotalProductCount(); err != nil {
		t.Fatal(err)
	} else {
		if want, got := 6, total; want != got {
			t.Errorf("Expected %d, got %d", want, got)
		}
	}
}
