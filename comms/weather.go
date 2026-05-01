package comms

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const weatherApiHost = "https://www.ncei.noaa.gov/cdo-web/api/v2/"

// LocationInfo represents weather station information
type LocationInfo struct {
	Id string
}

// GetStation retrieves weather station information for the given location
func (c *CommsClient) GetStation(location GeoCodeAPIResponseItem, apiToken string) (LocationInfo, error) {
	var l LocationInfo
	req, err := http.NewRequest(http.MethodGet, weatherApiHost+"stations", nil)
	if err != nil {
		return l, err
	}
	req.Header.Add("token", apiToken)
	q := url.Values{}
	q.Add("extent", fmt.Sprintf("%v,%v", location.Lat, location.Lon))
	req.URL.RawQuery = q.Encode()
	fmt.Println("req url", req.URL)
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
