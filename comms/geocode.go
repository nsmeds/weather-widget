package comms

import (
	"encoding/json"
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

// GetLocations retrieves locations for the given query
func (c *CommsClient) GetLocations(query string, apiKey string) ([]GeoCodeAPIResponseItem, error) {
	var l geoCodeAPIResponse
	// TODO possibly convert two-letter state code to three-letter, because API only handles the latter
	// handle spaces - API needs comma delimiter
	spaceToComma := strings.Join(strings.Split(query, " "), ",")
	req, err := http.NewRequest(http.MethodGet, geoCodingHost, nil)
	if err != nil {
		return l, err
	}
	q := url.Values{}
	q.Add("appid", apiKey)
	q.Add("q", spaceToComma)
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
	fmt.Printf("\nunmarshaled: %v\n", l)
	return l, nil
}
