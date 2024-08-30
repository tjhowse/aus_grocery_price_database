package coles

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"

	"github.com/tjhowse/aus_grocery_price_database/internal/utils"
)

const BROWSE_JSON_URL_FORMAT = "%s/_next/data/%s/en/browse.json"
const BROWSE_HOMEPAGE_URL_FORMAT = "%s/browse"
const CATEGORY_URL_FORMAT = "%s/_next/data/%s/en/browse/%s.json?slug=%s"

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
	} else {
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
func (c *Coles) getCategoryJSON(category string) ([]byte, error) {
	var req *http.Request
	var resp *http.Response
	var err error
	url := fmt.Sprintf(CATEGORY_URL_FORMAT, c.baseURL, c.colesAPIVersion, category, category)
	var body []byte

	fmt.Println("URL: ", url)

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
