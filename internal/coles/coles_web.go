package coles

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
)

const BROWSE_JSON_URL_FORMAT = "%s/_next/data/%s/en/browse.json"
const BROWSE_HOMEPAGE_URL_FORMAT = "%s/browse"

// get_url returns the URL for the Coles API.
func (c *Coles) get_url() string {
	return fmt.Sprintf(BROWSE_JSON_URL_FORMAT, c.baseURL, c.colesAPIVersion)
}

// updateAPIVersion grabs the coles home page and extracts the API version from it.
func (c *Coles) updateAPIVersion() error {
	// Get the browse homepage
	body, err := c.getBrowseHomepage()
	if err != nil {
		return fmt.Errorf("failed to get homepage: %w", err)
	}

	// Extract and update the API version
	if newAPI, err := extractAPIVersion(body); err != nil {
		return fmt.Errorf("failed to extract API version: %w", err)
	} else {
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
