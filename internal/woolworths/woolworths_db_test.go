package woolworths

import (
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/tjhowse/aus_grocery_price_database/internal/shared"
	utils "github.com/tjhowse/aus_grocery_price_database/internal/utils"
)

func TestUpdateProductInfo(t *testing.T) {
	w := getInitialisedWoolworths()
	testFile, err := utils.ReadEntireFile("data/category_1-E5BEE36E_1.json")
	if err != nil {
		t.Fatal(err)
	}
	infos, err := extractProductInfoFromProductListPage(testFile)
	if err != nil {
		t.Fatal(err)
	}

	infos[0].Updated = time.Now()

	if err := w.saveProductInfoNoTx(infos[0]); err != nil {
		t.Fatal(err)
	}

	var readProdInfo woolworthsProductInfo
	readProdInfo, err = w.loadProductInfo("133211")
	if err != nil {
		t.Fatal(err)
	}
	if readProdInfo.Info.Description != infos[0].Info.Description {
		t.Errorf("Expected %v, got %v", infos[0].Info.Description, readProdInfo.Info.Description)
	}
}

func TestMissingProduct(t *testing.T) {
	w := getInitialisedWoolworths()
	_, err := w.loadProductInfo("123456")
	if err == nil {
		t.Fatal("Expected an error")
	}
	if want, got := shared.ErrProductMissing, err; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
}

func TestDepartmentInfo(t *testing.T) {
	w := getInitialisedWoolworths()
	dept := departmentInfo{NodeID: "1-E5BEE36E", Description: "Fruit & Veg", Updated: time.Now()}
	w.saveDepartment(dept)
	departmentIDs, err := w.loadDepartmentInfoList()
	if err != nil {
		t.Fatal(err)
	}
	if want, got := 1, len(departmentIDs); want != got {
		t.Errorf("Expected %d, got %d", want, got)
	}
	if want, got := departmentID("1-E5BEE36E"), departmentIDs[0].NodeID; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
	if want, got := "Fruit & Veg", departmentIDs[0].Description; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
}

func TestDBFail(t *testing.T) {

	w := Woolworths{}
	err := w.Init("", "/zingabingo/db.db3", 5*time.Second)
	if err == nil {
		t.Fatal("Expected an error")
	}
	if want, got := "failed to create blank DB: unable to open database file: no such file or directory", err.Error(); want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
}

func TestGetSharedProductsUpdatedAfter(t *testing.T) {
	w := Woolworths{}
	w.Init(woolworthsServer.URL, ":memory:", 5*time.Second)
	w.filterDepartments = false
	var infoList []woolworthsProductInfo
	infoList = append(infoList, woolworthsProductInfo{ID: "123455", Info: productListPageProduct{DisplayName: "1", Price: decimal.NewFromFloat(1.5)}, Updated: time.Now().Add(-5 * time.Minute)})
	infoList = append(infoList, woolworthsProductInfo{ID: "123456", Info: productListPageProduct{DisplayName: "2", Price: decimal.NewFromFloat(2.4)}, Updated: time.Now().Add(-4 * time.Minute)})
	infoList = append(infoList, woolworthsProductInfo{ID: "123457", Info: productListPageProduct{DisplayName: "3", Price: decimal.NewFromFloat(3.3)}, Updated: time.Now().Add(-3 * time.Minute)})
	infoList = append(infoList, woolworthsProductInfo{ID: "123458", Info: productListPageProduct{DisplayName: "4", Price: decimal.NewFromFloat(4.2)}, Updated: time.Now().Add(-1 * time.Minute)})
	// Put this one in twice to test the PreviousPriceCents is updated.
	infoList = append(infoList, woolworthsProductInfo{ID: "123459", Info: productListPageProduct{DisplayName: "5", Price: decimal.NewFromFloat(5.0)}, Updated: time.Now()})
	infoList = append(infoList, woolworthsProductInfo{ID: "123459", Info: productListPageProduct{DisplayName: "5", Price: decimal.NewFromFloat(5.1)}, Updated: time.Now()})
	// This last one is to test that we don't get products that have a blank name.
	infoList = append(infoList, woolworthsProductInfo{ID: "123460", Info: productListPageProduct{DisplayName: "", Price: decimal.NewFromFloat(6.0)}, Updated: time.Now()})

	for _, info := range infoList {
		w.saveProductInfoNoTx(info)
	}
	productIDs, err := w.GetSharedProductsUpdatedAfter(time.Now().Add(-2*time.Minute), 10)
	if err != nil {
		t.Fatal(err)
	}
	if want, got := 2, len(productIDs); want != got {
		t.Fatalf("Expected %d products, got %d", want, got)
	}
	if want, got := WOOLWORTHS_ID_PREFIX+"123458", productIDs[0].ID; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
	if want, got := WOOLWORTHS_ID_PREFIX+"123459", productIDs[1].ID; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
	if want, got := 500, productIDs[1].PreviousPriceCents; want != got {
		t.Errorf("Expected %v, got %v", want, got)
	}
	if want, got := 510, productIDs[1].PriceCents; want != got {
		t.Errorf("Expected %v, got %v", want, got)
	}
	productIDs, err = w.GetSharedProductsUpdatedAfter(time.Now().Add(-2*time.Minute), 1)
	if err != nil {
		t.Fatal(err)
	}
	if want, got := 1, len(productIDs); want != got {
		t.Fatalf("Expected %d products, got %d", want, got)
	}
	if want, got := WOOLWORTHS_ID_PREFIX+"123458", productIDs[0].ID; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
	if want, got := 420, productIDs[0].PriceCents; want != got {
		t.Errorf("Expected %v, got %v", want, got)
	}

	if total, err := w.GetTotalProductCount(); err != nil {
		t.Fatal(err)
	} else {
		if want, got := 6, total; want != got {
			t.Errorf("Expected %d, got %d", want, got)
		}
	}
}

func TestSaveProductInfo(t *testing.T) {
	w := getInitialisedWoolworths()
	inProduct := woolworthsProductInfo{ID: "123455", Info: productListPageProduct{DisplayName: "1", Price: decimal.NewFromFloat(1.5)}, Updated: time.Now().Add(-5 * time.Minute)}

	err := w.saveProductInfoNoTx(inProduct)
	if err != nil {
		t.Fatal(err)
	}
	outProduct, err := w.loadProductInfo("123455")
	if err != nil {
		t.Fatal(err)
	}
	if want, got := inProduct.Info.Name, outProduct.Info.Name; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
	if want, got := inProduct.Info.Price.Mul(decimal.NewFromInt(100)), outProduct.Info.Price; want.Cmp(got) != 0 {
		t.Errorf("Expected %v, got %v", want, got)
	}
	if want, got := inProduct.departmentID, outProduct.departmentID; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
}

func TestPreviousPrice(t *testing.T) {
	w := getInitialisedWoolworths()
	inProduct := woolworthsProductInfo{ID: "123455", Info: productListPageProduct{DisplayName: "1", Price: decimal.NewFromFloat(1.5)}, Updated: time.Now().Add(-5 * time.Minute)}

	err := w.saveProductInfoNoTx(inProduct)
	if err != nil {
		t.Fatal(err)
	}
	outProduct, err := w.loadProductInfo("123455")
	if err != nil {
		t.Fatal(err)
	}
	if want, got := 0.0, outProduct.PreviousPrice; !got.Equal(decimal.NewFromFloat(want * 100)) {
		t.Errorf("Expected %v, got %v", want, got)
	}
	if want, got := 1.5, outProduct.Info.Price; !got.Equal(decimal.NewFromFloat(want * 100)) {
		t.Errorf("Expected %v, got %v", want, got)
	}
	inProduct = woolworthsProductInfo{ID: "123455", Info: productListPageProduct{DisplayName: "1", Price: decimal.NewFromFloat(2)}, Updated: time.Now().Add(-5 * time.Minute)}

	err = w.saveProductInfoNoTx(inProduct)
	if err != nil {
		t.Fatal(err)
	}
	outProduct, err = w.loadProductInfo("123455")
	if err != nil {
		t.Fatal(err)
	}
	if want, got := 2.0, outProduct.Info.Price; !got.Equal(decimal.NewFromFloat(want * 100)) {
		t.Errorf("Expected %v, got %v", want, got)
	}
	if want, got := 1.5, outProduct.PreviousPrice; !got.Equal(decimal.NewFromFloat(want * 100)) {
		t.Errorf("Expected %v, got %v", want, got)
	}
}

func TestBackupDB(t *testing.T) {

	// Get a temp directory

	tempDirName, err := os.MkdirTemp("", "delme")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDirName)

	func() {
		w := Woolworths{}
		err = w.Init(woolworthsServer.URL, tempDirName+"/delme.db3", 10*time.Minute)
		if err != nil {
			slog.Error("Failed to initialise Woolworths", "error", err)
		}
	}()

	func() {
		w := Woolworths{}
		err = w.Init(woolworthsServer.URL, tempDirName+"/delme.db3", 10*time.Minute)
		if err != nil {
			slog.Error("Failed to initialise Woolworths", "error", err)
		}
		matches, err := filepath.Glob(tempDirName + "/delme.db3.*")
		if err != nil {
			t.Fatal(err)
		}
		if want, got := 0, len(matches); want != got {
			t.Fatalf("Unexpectedly found a backup of the DB that shouldn't've been created.")
		}
		// Tweak the schema version to DB_SCHEMA_VERSION-1 to force a backup.
		w.db.Exec("UPDATE schema SET version = ?", DB_SCHEMA_VERSION-1)
	}()

	func() {
		w := Woolworths{}
		err = w.Init(woolworthsServer.URL, tempDirName+"/delme.db3", 10*time.Minute)
		if err != nil {
			slog.Error("Failed to initialise Woolworths", "error", err)
		}
		// Ensure we created a backup of the old database at tempDirName/delme.db3.{DB_SCHEMA_VERSION-1}.{timestamp}
		matches, err := filepath.Glob(tempDirName + "/delme.db3." + strconv.Itoa(DB_SCHEMA_VERSION-1) + ".*")
		if err != nil {
			t.Fatal(err)
		}
		if want, got := 1, len(matches); want != got {
			t.Fatalf("Couldn't find the backed-up file.")
		}
	}()

}
