package comms

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const cdoDataEndpoint = "https://www.ncei.noaa.gov/cdo-web/api/v2/data"

// ClimateNormals holds 30-year average high/low temperatures for a given day.
type ClimateNormals struct {
	TemperatureHighF float64 `json:"temperature_high_f"`
	TemperatureLowF  float64 `json:"temperature_low_f"`
	Period           string  `json:"period"`
}

type cdoDataResponse struct {
	Results []struct {
		Datatype string  `json:"datatype"`
		Value    float64 `json:"value"`
	} `json:"results"`
}

// GetClimateNormals fetches 30-year (1981-2010) daily climate normals for a
// GHCND station on a given calendar date. CDO NORMAL_DLY values for temperature
// are in tenths of degrees Fahrenheit.
func (c *CommsClient) GetClimateNormals(stationID string, date time.Time, apiToken string) (ClimateNormals, error) {
	var normals ClimateNormals

	// NORMAL_DLY dataset covers 1981-2010; use 2010 as the base year for the query date.
	// For Feb 29 (not present in 2010), fall back to Mar 1.
	month, day := date.Month(), date.Day()
	baseYear := 2010
	if month == time.February && day == 29 {
		month = time.March
		day = 1
	}
	queryDate := fmt.Sprintf("%d-%02d-%02d", baseYear, month, day)

	req, err := http.NewRequest(http.MethodGet, cdoDataEndpoint, nil)
	if err != nil {
		return normals, err
	}
	req.Header.Set("token", apiToken)
	q := url.Values{}
	q.Set("datasetid", "NORMAL_DLY")
	q.Set("stationid", stationID)
	q.Set("datatypeid", "DLY-TMAX-NORMAL,DLY-TMIN-NORMAL")
	q.Set("startdate", queryDate)
	q.Set("enddate", queryDate)
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return normals, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return normals, fmt.Errorf("CDO normals: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return normals, err
	}
	var parsed cdoDataResponse
	if err = json.Unmarshal(body, &parsed); err != nil {
		return normals, err
	}
	if len(parsed.Results) == 0 {
		return normals, errors.New("no climate normals available for this station and date")
	}
	for _, r := range parsed.Results {
		switch r.Datatype {
		case "DLY-TMAX-NORMAL":
			normals.TemperatureHighF = r.Value / 10
		case "DLY-TMIN-NORMAL":
			normals.TemperatureLowF = r.Value / 10
		}
	}
	normals.Period = "1981-2010 climate normals"
	return normals, nil
}
