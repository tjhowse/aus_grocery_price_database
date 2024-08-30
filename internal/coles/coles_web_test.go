package coles

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
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
		"data/fruit-vegetables.json",
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

		if strings.HasPrefix(r.URL.Path, "/browse") {
			responseFilename = "data/browse.html.file"
		} else if strings.HasPrefix(r.URL.Path, categoryPrefix) {
			var category string
			if _, err := fmt.Sscanf(r.URL.Path, categoryPrefix+"%s", &category); err != nil {
				slog.Error("Failed to parse category from URL", "error", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			responseFilename = fmt.Sprintf("data/%s", category)
		} else {
			w.WriteHeader(http.StatusNotFound)
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
	body, err := c.getCategoryJSON("fruit-vegetables")
	if err != nil {
		t.Fatalf("Failed to get category JSON: %v", err)
	}
	if len(body) == 0 {
		t.Fatalf("Got empty body")
	}
	// utils.WriteEntireFile("data/fruit-vegetables.json", body)
}
