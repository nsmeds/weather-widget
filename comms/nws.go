package comms

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const nwsAPIHost = "https://api.weather.gov"

// NWSPointsInfo contains the URLs needed to fetch weather data from NWS.
type NWSPointsInfo struct {
	ForecastURL            string
	ObservationStationsURL string
}

// CurrentConditions represents the latest weather observation at a station.
type CurrentConditions struct {
	TemperatureF float64 `json:"temperature_f"`
	Conditions   string  `json:"conditions"`
	Humidity     float64 `json:"humidity"`
	WindSpeedMph float64 `json:"wind_speed_mph"`
	Timestamp    string  `json:"timestamp"`
}

// TodayForecast represents the forecast for the current day.
type TodayForecast struct {
	Name             string  `json:"name"`
	ShortForecast    string  `json:"short_forecast"`
	TemperatureHighF float64 `json:"temperature_high_f"`
	TemperatureLowF  float64 `json:"temperature_low_f"`
}

type nwsPointsResponse struct {
	Properties struct {
		Forecast            string `json:"forecast"`
		ObservationStations string `json:"observationStations"`
	} `json:"properties"`
}

type nwsStationsResponse struct {
	Features []struct {
		Properties struct {
			StationIdentifier string `json:"stationIdentifier"`
		} `json:"properties"`
	} `json:"features"`
}

type nwsWindMeasurement struct {
	UnitCode string   `json:"unitCode"`
	Value    *float64 `json:"value"`
}

type nwsObservationResponse struct {
	Properties struct {
		Timestamp        string             `json:"timestamp"`
		TextDescription  string             `json:"textDescription"`
		Temperature      struct{ Value *float64 `json:"value"` } `json:"temperature"`
		RelativeHumidity struct{ Value *float64 `json:"value"` } `json:"relativeHumidity"`
		WindSpeed        nwsWindMeasurement `json:"windSpeed"`
	} `json:"properties"`
}

type nwsForecastResponse struct {
	Properties struct {
		Periods []struct {
			Name          string  `json:"name"`
			Temperature   float64 `json:"temperature"`
			IsDaytime     bool    `json:"isDaytime"`
			ShortForecast string  `json:"shortForecast"`
		} `json:"periods"`
	} `json:"properties"`
}

// GetNWSPoints fetches NWS grid point metadata for a lat/lon, returning forecast
// and observation station URLs.
func (c *CommsClient) GetNWSPoints(lat, lon float64) (NWSPointsInfo, error) {
	var info NWSPointsInfo
	url := fmt.Sprintf("%s/points/%.4f,%.4f", nwsAPIHost, lat, lon)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return info, err
	}
	req.Header.Set("User-Agent", "weather-widget/1.0")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return info, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return info, fmt.Errorf("NWS points: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return info, err
	}
	var parsed nwsPointsResponse
	if err = json.Unmarshal(body, &parsed); err != nil {
		return info, err
	}
	info.ForecastURL = parsed.Properties.Forecast
	info.ObservationStationsURL = parsed.Properties.ObservationStations
	return info, nil
}

// GetNWSObservationStation fetches the list of observation stations for a grid
// point and returns the identifier of the nearest one.
func (c *CommsClient) GetNWSObservationStation(stationsURL string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, stationsURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "weather-widget/1.0")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("NWS stations: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var parsed nwsStationsResponse
	if err = json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Features) == 0 {
		return "", errors.New("no observation stations found")
	}
	return parsed.Features[0].Properties.StationIdentifier, nil
}

// GetCurrentObservation fetches the latest observation from an NWS station.
// Temperature is converted from Celsius to Fahrenheit. Wind speed is converted
// to mph respecting the unitCode returned by the API (km_h-1 or m_s-1).
func (c *CommsClient) GetCurrentObservation(stationID string) (CurrentConditions, error) {
	var conditions CurrentConditions
	url := fmt.Sprintf("%s/stations/%s/observations/latest", nwsAPIHost, stationID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return conditions, err
	}
	req.Header.Set("User-Agent", "weather-widget/1.0")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return conditions, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return conditions, fmt.Errorf("NWS observation: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return conditions, err
	}
	var parsed nwsObservationResponse
	if err = json.Unmarshal(body, &parsed); err != nil {
		return conditions, err
	}
	p := parsed.Properties
	conditions.Conditions = p.TextDescription
	conditions.Timestamp = p.Timestamp
	if p.Temperature.Value != nil {
		conditions.TemperatureF = celsiusToFahrenheit(*p.Temperature.Value)
	}
	if p.RelativeHumidity.Value != nil {
		conditions.Humidity = *p.RelativeHumidity.Value
	}
	if p.WindSpeed.Value != nil {
		conditions.WindSpeedMph = windToMph(p.WindSpeed.UnitCode, *p.WindSpeed.Value)
	}
	return conditions, nil
}

// GetNWSForecast fetches the forecast for the current period from the NWS
// forecast URL. It returns the daytime high and overnight low for today.
func (c *CommsClient) GetNWSForecast(forecastURL string) (TodayForecast, error) {
	var forecast TodayForecast
	req, err := http.NewRequest(http.MethodGet, forecastURL, nil)
	if err != nil {
		return forecast, err
	}
	req.Header.Set("User-Agent", "weather-widget/1.0")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return forecast, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return forecast, fmt.Errorf("NWS forecast: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return forecast, err
	}
	var parsed nwsForecastResponse
	if err = json.Unmarshal(body, &parsed); err != nil {
		return forecast, err
	}
	periods := parsed.Properties.Periods
	if len(periods) == 0 {
		return forecast, errors.New("no forecast periods available")
	}
	var highFound, lowFound bool
	for _, p := range periods {
		if p.IsDaytime && !highFound {
			forecast.Name = p.Name
			forecast.ShortForecast = p.ShortForecast
			forecast.TemperatureHighF = p.Temperature
			highFound = true
		}
		if !p.IsDaytime && !lowFound {
			forecast.TemperatureLowF = p.Temperature
			lowFound = true
		}
		if highFound && lowFound {
			break
		}
	}
	return forecast, nil
}

func celsiusToFahrenheit(c float64) float64 {
	return c*9/5 + 32
}

func kmhToMph(kmh float64) float64 {
	return kmh * 0.621371
}

// windToMph converts wind speed to mph, respecting the NWS unitCode.
// NWS observations use wmoUnit:km_h-1; wmoUnit:m_s-1 is handled defensively.
func windToMph(unitCode string, value float64) float64 {
	switch unitCode {
	case "wmoUnit:km_h-1":
		return kmhToMph(value)
	case "wmoUnit:m_s-1":
		return value * 2.236936
	default:
		fmt.Printf("windToMph: unknown unitCode %q, treating as km/h\n", unitCode)
		return kmhToMph(value)
	}
}
