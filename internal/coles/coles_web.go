package coles

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"

	"github.com/tjhowse/aus_grocery_price_database/internal/utils"
)

const DEFAULT_API_VERSION = "20240827.02_v4.7.7"

const BROWSE_JSON_URL_FORMAT = "%s/_next/data/%s/en/browse.json"
const BROWSE_HOMEPAGE_URL_FORMAT = "%s/browse"
const CATEGORY_URL_FORMAT = "%s/_next/data/%s/en/browse/%s.json"
const SCRAPE_TRAP_STRING = "Pardon Our Interruption"

var ErrHitScrapeTrap = errors.New("caught in a scrape trap")

// updateAPIVersion grabs the coles home page and extracts the API version from it.
func (c *Coles) updateAPIVersion() error {
	// Get the browse homepage
	body, err := c.getBrowseHomepage()
	if err != nil {
		return fmt.Errorf("failed to get homepage: %w", err)
	}

	// Extract and update the API version
	if newAPI, err := extractAPIVersion(body); err != nil {

		if err := utils.WriteEntireFile("failed_coles_homepage.html", body); err != nil {
			slog.Error("Failed to write failed homepage to file", "error", err)
		}
		return fmt.Errorf("failed to extract API version: %w", err)
	} else if newAPI != c.colesAPIVersion {
		slog.Info("Updated API version", "old_version", c.colesAPIVersion, "version", newAPI)
		c.colesAPIVersion = newAPI
	}
	return nil
}

// getBrowseHomepage returns the bytes of the Coles browse homepage.
func (c *Coles) getBrowseHomepage() ([]byte, error) {
	var req *http.Request
	var resp *http.Response
	var err error
	url := fmt.Sprintf(BROWSE_HOMEPAGE_URL_FORMAT, c.baseURL)
	var body []byte

	if req, err = http.NewRequest("GET", url, nil); err != nil {
		return body, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:129.0) Gecko/20100101 Firefox/129.0")
	req.Header.Set("Accept", "text/html")

	resp, err = c.client.Do(req)
	if err != nil {
		return body, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return body, fmt.Errorf("failed to get category data: %s", resp.Status)
	}
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return body, err
	}
	if checkForScrapeTrap(body) {
		return body, ErrHitScrapeTrap
	}
	return body, nil
}

// extractAPIVersion extracts the API version from the given HTML.
func extractAPIVersion(body []byte) (string, error) {
	// This locates the string ',"buildId":"20240827.02_v4.7.7",' in the body of the html and extracts
	// the '20240827.02_v4.7.7' value.

	r := regexp.MustCompile(`,"buildId":"(\d{8}\.\d{2}_v\d\.\d\.\d)",`)
	matches := r.FindSubmatch(body)
	if len(matches) != 2 {
		return "", fmt.Errorf("failed to find API version in body")
	}
	return string(matches[1]), nil
}

// getBrowseJSON returns the bytes of the Coles browse JSON.
func (c *Coles) getBrowseJSON() ([]byte, error) {
	var req *http.Request
	var resp *http.Response
	var err error
	url := fmt.Sprintf(BROWSE_JSON_URL_FORMAT, c.baseURL, c.colesAPIVersion)
	var body []byte

	if req, err = http.NewRequest("GET", url, nil); err != nil {
		return body, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:129.0) Gecko/20100101 Firefox/129.0")
	req.Header.Set("Accept", "application/json")

	resp, err = c.client.Do(req)
	if err != nil {
		return body, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return body, fmt.Errorf("failed to get category data: %s", resp.Status)
	}
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return body, err
	}
	return body, nil
}

// getCategoryJSON returns the bytes of the Coles category JSON.
func (c *Coles) getCategoryJSON(category string, page int) ([]byte, error) {
	var req *http.Request
	var resp *http.Response
	var err error
	url := fmt.Sprintf(CATEGORY_URL_FORMAT, c.baseURL, c.colesAPIVersion, category)
	var body []byte

	if req, err = http.NewRequest("GET", url, nil); err != nil {
		return body, err
	}
	q := req.URL.Query()
	q.Add("slug", category)
	q.Add("page", strconv.Itoa(page))
	req.URL.RawQuery = q.Encode()
	// = req.URL.Query().Add("slug", category)
	fmt.Println(req.URL)

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:129.0) Gecko/20100101 Firefox/129.0")
	req.Header.Set("Accept", "application/json")

	resp, err = c.client.Do(req)
	if err != nil {
		return body, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return body, fmt.Errorf("failed to get category data: %s", resp.Status)
	}
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return body, err
	}
	return body, nil
}

// checkForScrapeTrap checks the given body for a scrape trap.
func checkForScrapeTrap(body []byte) bool {
	return bytes.Contains(body, []byte(SCRAPE_TRAP_STRING))
}

// getCategoryContents fetches a category page from the Coles API and unmarshals it.
func (c *Coles) getCategoryContents(category string, page int) (categoryPage, error) {
	body, err := c.getCategoryJSON(category, page)
	if err != nil {
		return categoryPage{}, err
	}
	// Unmarshal into a categoryPage
	var catPage categoryPage
	err = json.Unmarshal(body, &catPage)
	if err != nil {
		return categoryPage{}, fmt.Errorf("failed to unmarshal category page: %w", err)
	}
	return catPage, nil
}

// getProductsAndTotalCountForCategoryPage fetches the specified page of the specified category
// and returns the products and the total count of products in the category.
func (c *Coles) getProductsAndTotalCountForCategoryPage(dp departmentPage) ([]colesProductInfo, int, error) {
	catPage, err := c.getCategoryContents(dp.ID, dp.page)
	if err != nil {
		return nil, 0, err
	}
	// Filter out products without "_type" == "PRODUCT"
	var products []colesProductInfo
	for _, result := range catPage.PageProps.SearchResults.Results {
		if result.Type == "PRODUCT" {
			var product colesProductInfo
			product.Info = result
			product.RawJSON, err = json.Marshal(result)
			if err != nil {
				slog.Warn("Failed to marshal product info for storage", "error", err)
			}
			product.departmentID = dp.ID
			products = append(products, product)
		}
	}
	return products, catPage.PageProps.SearchResults.NoOfResults, nil
}

func (c *Coles) getDepartmentInfos() ([]departmentInfo, error) {
	body, err := c.getBrowseJSON()
	if err != nil {
		return nil, err
	}
	var browseJSON browsePage
	err = json.Unmarshal(body, &browseJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal browse JSON: %w", err)
	}
	return browseJSON.PageProps.AllProductCategories.CatalogGroupView, nil
}
