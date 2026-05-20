package comms_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/nsmeds/weather-widget/comms"
)

func TestGetNWSPointsSuccess(t *testing.T) {
	mockHTTPClient := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(strings.NewReader(`{
					"properties": {
						"forecast": "https://api.weather.gov/gridpoints/BOX/71,90/forecast",
						"observationStations": "https://api.weather.gov/gridpoints/BOX/71,90/stations"
					}
				}`)),
			}, nil
		},
	}

	client := comms.NewClient(mockHTTPClient)
	info, err := client.GetNWSPoints(42.36, -71.06)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if info.ForecastURL != "https://api.weather.gov/gridpoints/BOX/71,90/forecast" {
		t.Errorf("unexpected ForecastURL: %q", info.ForecastURL)
	}
	if info.ObservationStationsURL != "https://api.weather.gov/gridpoints/BOX/71,90/stations" {
		t.Errorf("unexpected ObservationStationsURL: %q", info.ObservationStationsURL)
	}
}

func TestGetNWSPointsAPIError(t *testing.T) {
	mockHTTPClient := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader(`{"title": "Not Found"}`)),
			}, nil
		},
	}

	client := comms.NewClient(mockHTTPClient)
	_, err := client.GetNWSPoints(0, 0)
	if err == nil {
		t.Error("expected error for non-200 response, got nil")
	}
}

func TestGetNWSPointsNetworkError(t *testing.T) {
	mockHTTPClient := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return nil, io.EOF
		},
	}

	client := comms.NewClient(mockHTTPClient)
	_, err := client.GetNWSPoints(42.36, -71.06)
	if err == nil {
		t.Error("expected error for network failure, got nil")
	}
}

func TestGetNWSObservationStationSuccess(t *testing.T) {
	mockHTTPClient := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(strings.NewReader(`{
					"features": [
						{"properties": {"stationIdentifier": "KBOS"}},
						{"properties": {"stationIdentifier": "KORD"}}
					]
				}`)),
			}, nil
		},
	}

	client := comms.NewClient(mockHTTPClient)
	stationID, err := client.GetNWSObservationStation("https://api.weather.gov/gridpoints/BOX/71,90/stations")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if stationID != "KBOS" {
		t.Errorf("expected first station 'KBOS', got %q", stationID)
	}
}

func TestGetNWSObservationStationEmpty(t *testing.T) {
	mockHTTPClient := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"features": []}`)),
			}, nil
		},
	}

	client := comms.NewClient(mockHTTPClient)
	_, err := client.GetNWSObservationStation("https://api.weather.gov/gridpoints/BOX/71,90/stations")
	if err == nil {
		t.Error("expected error when no stations returned, got nil")
	}
}

func TestGetNWSObservationStationAPIError(t *testing.T) {
	mockHTTPClient := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader(`{}`)),
			}, nil
		},
	}

	client := comms.NewClient(mockHTTPClient)
	_, err := client.GetNWSObservationStation("https://api.weather.gov/gridpoints/BOX/71,90/stations")
	if err == nil {
		t.Error("expected error for non-200 response, got nil")
	}
}

func TestGetCurrentObservationKmhWind(t *testing.T) {
	mockHTTPClient := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(strings.NewReader(`{
					"properties": {
						"timestamp": "2026-05-07T14:00:00+00:00",
						"textDescription": "Partly Cloudy",
						"temperature": {"value": 18.33},
						"relativeHumidity": {"value": 72.5},
						"windSpeed": {"unitCode": "wmoUnit:km_h-1", "value": 14.8}
					}
				}`)),
			}, nil
		},
	}

	client := comms.NewClient(mockHTTPClient)
	obs, err := client.GetCurrentObservation("KBOS")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if obs.Conditions != "Partly Cloudy" {
		t.Errorf("expected 'Partly Cloudy', got %q", obs.Conditions)
	}
	// 18.33°C → ~64.99°F
	if obs.TemperatureF < 64 || obs.TemperatureF > 66 {
		t.Errorf("expected ~65°F, got %.2f", obs.TemperatureF)
	}
	// 14.8 km/h → ~9.2 mph
	if obs.WindSpeedMph < 9 || obs.WindSpeedMph > 10 {
		t.Errorf("expected ~9.2 mph, got %.2f", obs.WindSpeedMph)
	}
	if obs.Timestamp != "2026-05-07T14:00:00+00:00" {
		t.Errorf("unexpected timestamp: %q", obs.Timestamp)
	}
}

func TestGetCurrentObservationMsWind(t *testing.T) {
	mockHTTPClient := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(strings.NewReader(`{
					"properties": {
						"timestamp": "2026-05-07T14:00:00+00:00",
						"textDescription": "Clear",
						"temperature": {"value": 20.0},
						"relativeHumidity": {"value": 50.0},
						"windSpeed": {"unitCode": "wmoUnit:m_s-1", "value": 4.0}
					}
				}`)),
			}, nil
		},
	}

	client := comms.NewClient(mockHTTPClient)
	obs, err := client.GetCurrentObservation("KBOS")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// 4.0 m/s → ~8.95 mph
	if obs.WindSpeedMph < 8.9 || obs.WindSpeedMph > 9.0 {
		t.Errorf("expected ~8.95 mph for 4 m/s, got %.2f", obs.WindSpeedMph)
	}
}

func TestGetCurrentObservationNullValues(t *testing.T) {
	mockHTTPClient := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(strings.NewReader(`{
					"properties": {
						"timestamp": "2026-05-07T14:00:00+00:00",
						"textDescription": "Unknown",
						"temperature": {"value": null},
						"relativeHumidity": {"value": null},
						"windSpeed": {"value": null}
					}
				}`)),
			}, nil
		},
	}

	client := comms.NewClient(mockHTTPClient)
	obs, err := client.GetCurrentObservation("KBOS")
	if err != nil {
		t.Fatalf("expected no error for null values, got %v", err)
	}
	if obs.TemperatureF != 0 {
		t.Errorf("expected 0 for null temperature, got %v", obs.TemperatureF)
	}
}

func TestGetCurrentObservationAPIError(t *testing.T) {
	mockHTTPClient := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader(`{}`)),
			}, nil
		},
	}

	client := comms.NewClient(mockHTTPClient)
	_, err := client.GetCurrentObservation("KBOS")
	if err == nil {
		t.Error("expected error for non-200 response, got nil")
	}
}

func TestGetNWSForecastSuccess(t *testing.T) {
	mockHTTPClient := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(strings.NewReader(`{
					"properties": {
						"periods": [
							{
								"name": "Today",
								"temperature": 68,
								"temperatureUnit": "F",
								"isDaytime": true,
								"shortForecast": "Mostly Sunny"
							},
							{
								"name": "Tonight",
								"temperature": 45,
								"temperatureUnit": "F",
								"isDaytime": false,
								"shortForecast": "Partly Cloudy"
							}
						]
					}
				}`)),
			}, nil
		},
	}

	client := comms.NewClient(mockHTTPClient)
	forecast, err := client.GetNWSForecast("https://api.weather.gov/gridpoints/BOX/71,90/forecast")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if forecast.Name != "Today" {
		t.Errorf("expected name 'Today', got %q", forecast.Name)
	}
	if forecast.ShortForecast != "Mostly Sunny" {
		t.Errorf("expected 'Mostly Sunny', got %q", forecast.ShortForecast)
	}
	if forecast.TemperatureHighF != 68 {
		t.Errorf("expected high 68, got %v", forecast.TemperatureHighF)
	}
	if forecast.TemperatureLowF != 45 {
		t.Errorf("expected low 45, got %v", forecast.TemperatureLowF)
	}
}

func TestGetNWSForecastNightFirst(t *testing.T) {
	// When the first period is nighttime (late evening request), high comes from the next day
	mockHTTPClient := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(strings.NewReader(`{
					"properties": {
						"periods": [
							{
								"name": "Tonight",
								"temperature": 38,
								"isDaytime": false,
								"shortForecast": "Clear"
							},
							{
								"name": "Thursday",
								"temperature": 62,
								"isDaytime": true,
								"shortForecast": "Sunny"
							}
						]
					}
				}`)),
			}, nil
		},
	}

	client := comms.NewClient(mockHTTPClient)
	forecast, err := client.GetNWSForecast("https://api.weather.gov/gridpoints/BOX/71,90/forecast")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if forecast.TemperatureLowF != 38 {
		t.Errorf("expected low 38, got %v", forecast.TemperatureLowF)
	}
	if forecast.TemperatureHighF != 62 {
		t.Errorf("expected high 62, got %v", forecast.TemperatureHighF)
	}
}

func TestGetNWSForecastAPIError(t *testing.T) {
	mockHTTPClient := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusServiceUnavailable,
				Body:       io.NopCloser(strings.NewReader(`{}`)),
			}, nil
		},
	}

	client := comms.NewClient(mockHTTPClient)
	_, err := client.GetNWSForecast("https://api.weather.gov/gridpoints/BOX/71,90/forecast")
	if err == nil {
		t.Error("expected error for non-200 response, got nil")
	}
}

func TestGetNWSForecastZeroFahrenheitLow(t *testing.T) {
	// 0°F is a valid temperature; the zero sentinel bug would have dropped it
	mockHTTPClient := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(strings.NewReader(`{
					"properties": {
						"periods": [
							{"name": "Today", "temperature": 15, "isDaytime": true, "shortForecast": "Sunny"},
							{"name": "Tonight", "temperature": 0, "isDaytime": false, "shortForecast": "Clear"}
						]
					}
				}`)),
			}, nil
		},
	}

	client := comms.NewClient(mockHTTPClient)
	forecast, err := client.GetNWSForecast("https://api.weather.gov/gridpoints/BOX/71,90/forecast")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if forecast.TemperatureLowF != 0 {
		t.Errorf("expected low exactly 0°F, got %v", forecast.TemperatureLowF)
	}
	if forecast.TemperatureHighF != 15 {
		t.Errorf("expected high 15, got %v", forecast.TemperatureHighF)
	}
}

func TestGetNWSForecastNoPeriods(t *testing.T) {
	mockHTTPClient := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"properties": {"periods": []}}`)),
			}, nil
		},
	}

	client := comms.NewClient(mockHTTPClient)
	_, err := client.GetNWSForecast("https://api.weather.gov/gridpoints/BOX/71,90/forecast")
	if err == nil {
		t.Error("expected error when no forecast periods returned, got nil")
	}
}
