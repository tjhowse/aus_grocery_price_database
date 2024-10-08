package coles

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tjhowse/aus_grocery_price_database/internal/utils"
)

var colesServer = ColesHTTPServer()

func getInitialisedColes() Coles {
	c := Coles{}
	err := c.Init(colesServer.URL, ":memory:", 10*time.Minute)
	if err != nil {
		slog.Error("Failed to initialise Coles", "error", err)
	}
	return c
}

// This mocks enough of the Woolworths API to test various stuff
func ColesHTTPServer() *httptest.Server {
	var err error

	filesToLoad := []string{
		"data/browse.json",
		"data/browse.html.file",
		"data/fruit-vegetables_1.json",
		"data/fruit-vegetables_2.json",
	}
	fileContents := make(map[string][]byte)
	for _, filename := range filesToLoad {
		fileContents[filename], err = utils.ReadEntireFile(filename)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to read file %s: %v\n", filename, err))
		}
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var responseFilename string

		categoryPrefix := fmt.Sprintf("/_next/data/%s/en/browse/", DEFAULT_API_VERSION)
		browseJSONPrefix := fmt.Sprintf("/_next/data/%s/en/browse.json", DEFAULT_API_VERSION)

		if strings.HasPrefix(r.URL.Path, "/browse") {
			responseFilename = "data/browse.html.file"
		} else if strings.HasPrefix(r.URL.Path, browseJSONPrefix) {
			responseFilename = "data/browse.json"
		} else if strings.HasPrefix(r.URL.Path, categoryPrefix) {
			var category string
			var page int
			if _, err := fmt.Sscanf(r.URL.Path, categoryPrefix+"%s", &category); err != nil {
				slog.Error("Failed to parse category from URL", "error", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if pageStr := r.URL.Query().Get("page"); pageStr == "" {
				page = 1
			} else if page, err = strconv.Atoi(pageStr); err != nil {
				slog.Error("Failed to parse page from URL", "error", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			category = strings.TrimSuffix(category, ".json")
			responseFilename = fmt.Sprintf("data/%s_%d.json", category, page)
		} else {
			w.WriteHeader(http.StatusNotFound)
			slog.Error("Simulated woolworths server can't find requested URL.", "url", r.URL.Path)
			return
		}

		if responseData, knownFile := fileContents[responseFilename]; !knownFile {
			slog.Error("Simulated woolworths server can't find requested file.", "filename", responseFilename)
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(responseData)
		}
	}))
}

func TestGetHomepage(t *testing.T) {
	c := getInitialisedColes()
	body, err := c.getBrowseHomepage()
	if err != nil {
		t.Errorf("Failed to get homepage: %v", err)
	}
	if len(body) == 0 {
		t.Errorf("Got empty body")
	}
}

func TestExtractAPIVersion(t *testing.T) {
	body, err := utils.ReadEntireFile("data/browse.html.file")
	if body == nil || err != nil {
		t.Errorf("Failed to read file")
	}
	if version, err := extractAPIVersion(body); err != nil {
		t.Errorf("Failed to extract API version: %v", err)
	} else {
		if want, got := "20240827.02_v4.7.7", version; want != got {
			t.Errorf("Expected %s, got %s", want, got)
		}
	}
}

func TestExtractAPIVersionAgain(t *testing.T) {
	body, err := utils.ReadEntireFile("data/browse_different_api_version.html.file")
	if body == nil || err != nil {
		t.Errorf("Failed to read file")
	}
	if version, err := extractAPIVersion(body); err != nil {
		t.Errorf("Failed to extract API version: %v", err)
	} else {
		if want, got := "20240904.01_v4.10.0", version; want != got {
			t.Errorf("Expected %s, got %s", want, got)
		}
	}
}

func TestCheckForScrapeTrap(t *testing.T) {
	body, err := utils.ReadEntireFile("data/scrape_trap.html.file")
	if body == nil || err != nil {
		t.Errorf("Failed to read file")
	}
	if !checkForScrapeTrap(body) {
		t.Errorf("Failed to detect scrape trap")
	}
}

func TestUpdateAPIVersion(t *testing.T) {
	c := getInitialisedColes()
	// Set a deliberately old version
	c.colesAPIVersion = "20240809.03_v4.7.3"
	err := c.updateAPIVersion()
	if err != nil {
		t.Errorf("Failed to update API version: %v", err)
	}
	if want, got := "20240827.02_v4.7.7", c.colesAPIVersion; want != got {
		t.Errorf("Expected %s, got %s", want, got)
	}
}

func TestGetCategoryJSON(t *testing.T) {
	c := getInitialisedColes()
	// c.baseURL = "https://coles.com.au"
	// c.colesAPIVersion = "20240827.02_v4.7.7"
	// if err := c.updateAPIVersion(); err != nil {
	// 	t.Fatalf("Failed to update API version: %v", err)
	// }
	body, err := c.getCategoryJSON("fruit-vegetables", 1)
	if err != nil {
		t.Fatalf("Failed to get category JSON: %v", err)
	}
	if len(body) == 0 {
		t.Fatalf("Got empty body")
	}
	// utils.WriteEntireFile("data/fruit-vegetables.json", body)
}

func TestGetProductsAndTotalCountForCategoryPage(t *testing.T) {
	c := getInitialisedColes()

	{
		dp := departmentPage{"fruit-vegetables", 1}
		products, totalRecordCount, err := c.getProductsAndTotalCountForCategoryPage(dp)
		if err != nil {
			t.Fatalf("Failed to get products: %v", err)
		}
		if want, got := 48, len(products); want != got {
			t.Errorf("Expected %d products, got %d", want, got)
		}
		if want, got := "Bananas Mini Pack", products[0].Info.Name; want != got {
			t.Errorf("Expected %s, got %s", want, got)
		}
		if want, got := productID("2511791"), products[0].ID; want != got {
			t.Errorf("Expected %s, got %s", want, got)
		}
		if want, got := 578, totalRecordCount; want != got {
			t.Errorf("Expected %d total record count, got %d", want, got)
		}

	}
	{
		dp := departmentPage{"fruit-vegetables", 2}
		products, totalRecordCount, err := c.getProductsAndTotalCountForCategoryPage(dp)
		if err != nil {
			t.Fatalf("Failed to get products: %v", err)
		}
		if want, got := 2, len(products); want != got {
			t.Errorf("Expected %d products, got %d", want, got)
		}
		if want, got := "Glorba beans", products[0].Info.Name; want != got {
			t.Errorf("Expected %s, got %s", want, got)
		}
		if want, got := productID("8888888"), products[0].ID; want != got {
			t.Errorf("Expected %s, got %s", want, got)
		}
		if want, got := 50, totalRecordCount; want != got {
			t.Errorf("Expected %d total record count, got %d", want, got)
		}
	}
}

func TestGetDepartmentInfos(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	// c := getInitialisedRealColes()
	c := getInitialisedColes()
	// c.updateAPIVersion()

	departments, err := c.getDepartmentInfos()
	if err != nil {
		t.Fatalf("Failed to get department list: %v", err)
	}
	if want, got := 16, len(departments); want != got {
		t.Errorf("Expected %d departments, got %d", want, got)
	}
	var testCases = []struct {
		seoToken string
		name     string
		index    int
	}{
		{"fruit-vegetables", "Fruit & Vegetables", 3},
		{"meat-seafood", "Meat & Seafood", 2},
	}
	for _, tc := range testCases {
		if want, got := tc.seoToken, departments[tc.index].SeoToken; want != got {
			t.Errorf("Expected %s, got %s", want, got)
		}
		if want, got := tc.name, departments[tc.index].Name; want != got {
			t.Errorf("Expected %s, got %s", want, got)
		}
	}
}
