package comms

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const weatherApiHost = "https://www.ncei.noaa.gov/cdo-web/api/v2/"

// StationInfo represents a NOAA CDO/GHCND weather station.
type StationInfo struct {
	Id string
}

type cdoStationsResponse struct {
	Results []struct {
		Id string `json:"id"`
	} `json:"results"`
}

// GetStation retrieves the nearest GHCND weather station for the given location.
func (c *CommsClient) GetStation(location GeoCodeAPIResponseItem, apiToken string) (StationInfo, error) {
	var info StationInfo
	req, err := http.NewRequest(http.MethodGet, weatherApiHost+"stations", nil)
	if err != nil {
		return info, err
	}
	req.Header.Add("token", apiToken)
	q := url.Values{}
	// extent is a bounding box: minLat,minLon,maxLat,maxLon
	q.Add("extent", fmt.Sprintf("%.4f,%.4f,%.4f,%.4f",
		location.Lat-0.5, location.Lon-0.5,
		location.Lat+0.5, location.Lon+0.5,
	))
	q.Add("datasetid", "GHCND")
	q.Add("limit", "1")
	req.URL.RawQuery = q.Encode()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return info, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return info, fmt.Errorf("CDO stations: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return info, err
	}
	var parsed cdoStationsResponse
	if err = json.Unmarshal(body, &parsed); err != nil {
		return info, err
	}
	if len(parsed.Results) == 0 {
		return info, errors.New("no weather stations found near location")
	}
	info.Id = parsed.Results[0].Id
	return info, nil
}
