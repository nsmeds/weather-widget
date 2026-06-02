package comms

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// For converting zip code to city/state:
// https://github.com/USPS/api-examples?tab=readme-ov-file#city-and-state
// curl	-X 'GET' 'https://api.usps.com/addresses/v1/city-state?ZIPCode=30022' \
// 	--header 'accept: application/json' \
// 	--header 'X-User-Id: XXXXXXXXXXX' \
// 	--header 'Authorization: Bearer $TOKEN' \

// or just use this api
// https://openweathermap.org/api/geocoding-api

const geoCodingHost = "https://api.openweathermap.org/geo/1.0/direct"
const isUnitedStatesOnly = true

// GeoCodeAPIResponseItem represents a single geocoding API response
type GeoCodeAPIResponseItem struct {
	Name    string
	Lat     float64
	Lon     float64
	Country string
	State   string
}

type geoCodeAPIResponse []GeoCodeAPIResponseItem

// HTTPClient interface allows for dependency injection of HTTP clients
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var defaultHTTPClient = &http.Client{
	Timeout: time.Second * 5,
}

// CommsClient handles communication with external APIs for weather and location data
type CommsClient struct {
	httpClient HTTPClient
}

// NewClient creates a new CommsClient with the provided HTTP client
func NewClient(httpClient HTTPClient) *CommsClient {
	return &CommsClient{
		httpClient: httpClient,
	}
}

// NewClientWithDefaults creates a new CommsClient with default HTTP client settings
func NewClientWithDefaults() *CommsClient {
	return &CommsClient{
		httpClient: defaultHTTPClient,
	}
}

// normalizeGeoQuery converts a free-form "City, State" or "City State" query
// into the comma-separated form OpenWeatherMap requires, appending ",US" so
// that 2-letter state abbreviations are resolved correctly. Without the country
// code the API returns an empty result for US state abbreviations.
func normalizeGeoQuery(query string) (string, error) {
	// Split on commas and/or spaces so "Boston, MA", "Boston MA", and "Boston,MA"
	// all produce the same token list.
	const USCountryPrefix = "US"
	fields := strings.FieldsFunc(strings.TrimSpace(query), func(r rune) bool {
		return r == ',' || r == ' '
	})
	if isUnitedStatesOnly {
		return strings.Join(fields, ",") + "," + USCountryPrefix, nil
	}
	return "", errors.New("no country code configured; geographic scope must be set before querying")
}

// GetLocations retrieves locations for the given query
func (c *CommsClient) GetLocations(query string, apiKey string) ([]GeoCodeAPIResponseItem, error) {
	var l geoCodeAPIResponse
	req, err := http.NewRequest(http.MethodGet, geoCodingHost, nil)
	if err != nil {
		return l, err
	}
	normalized, err := normalizeGeoQuery(query)
	if err != nil {
		return l, err
	}
	q := url.Values{}
	q.Add("appid", apiKey)
	q.Add("q", normalized)
	req.URL.RawQuery = q.Encode()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return l, err
	}
	if resp.StatusCode != http.StatusOK {
		return l, fmt.Errorf("status code not ok: %v", resp.StatusCode)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return l, err
	}
	if err = json.Unmarshal(body, &l); err != nil {
		return l, err
	}
	return l, nil
}
