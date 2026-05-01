package comms_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/nsmeds/weather-widget/comms"
)

// mockClient is a test double for http.Client
type mockClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

// TestGetLocationsSuccess tests GetLocations with a successful response
func TestGetLocationsSuccess(t *testing.T) {
	mockHTTPClient := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(strings.NewReader(`[
					{
						"name": "New Orleans",
						"lat": 29.951065,
						"lon": -90.071533,
						"country": "US",
						"state": "NY"
					}
				]`)),
			}, nil
		},
	}

	client := comms.NewClient(mockHTTPClient)
	locations, err := client.GetLocations("New Orleans", "test-key")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(locations) == 0 {
		t.Error("expected locations, got empty response")
	}
	if len(locations) > 0 && locations[0].Name != "New Orleans" {
		t.Errorf("expected name 'New Orleans', got %q", locations[0].Name)
	}
}

// TestGetLocationsAPIError tests GetLocations with an API error response
func TestGetLocationsAPIError(t *testing.T) {
	mockHTTPClient := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       io.NopCloser(strings.NewReader(`{"message": "unauthorized"}`)),
			}, nil
		},
	}

	client := comms.NewClient(mockHTTPClient)
	locations, err := client.GetLocations("New Orleans", "invalid-key")
	if err == nil {
		t.Error("expected error for unauthorized response, got nil")
	}
	if len(locations) > 0 {
		t.Error("expected empty locations on error")
	}
}

// TestGetLocationsNetworkError tests GetLocations with a network error
func TestGetLocationsNetworkError(t *testing.T) {
	mockHTTPClient := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return nil, io.EOF
		},
	}

	client := comms.NewClient(mockHTTPClient)
	locations, err := client.GetLocations("New Orleans", "test-key")
	if err == nil {
		t.Error("expected error for network failure, got nil")
	}
	if locations != nil {
		t.Error("expected nil locations on error")
	}
}

// TestGetLocationsInvalidJSON tests GetLocations with invalid JSON response
func TestGetLocationsInvalidJSON(t *testing.T) {
	mockHTTPClient := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{invalid json}`)),
			}, nil
		},
	}

	client := comms.NewClient(mockHTTPClient)
	locations, err := client.GetLocations("New Orleans", "test-key")
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
	if locations != nil {
		t.Error("expected nil locations on parse error")
	}
}

// TestGetStationSuccess tests GetStation with a successful response
func TestGetStationSuccess(t *testing.T) {
	mockHTTPClient := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(strings.NewReader(`{
					"Id": "STATION123",
					"results": []
				}`)),
			}, nil
		},
	}

	location := comms.GeoCodeAPIResponseItem{
		Name: "New Orleans",
		Lat:  29.951065,
		Lon:  -90.071533,
	}

	client := comms.NewClient(mockHTTPClient)
	station, err := client.GetStation(location, "test-token")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if station.Id != "STATION123" {
		t.Errorf("expected station ID 'STATION123', got %q", station.Id)
	}
}

// TestGetStationAPIError tests GetStation with an API error
func TestGetStationAPIError(t *testing.T) {
	mockHTTPClient := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader(`{"error": "server error"}`)),
			}, nil
		},
	}

	location := comms.GeoCodeAPIResponseItem{
		Name: "New Orleans",
		Lat:  29.951065,
		Lon:  -90.071533,
	}

	client := comms.NewClient(mockHTTPClient)
	station, err := client.GetStation(location, "test-token")
	if err == nil {
		t.Error("expected error for server error, got nil")
	}
	if station.Id != "" {
		t.Errorf("expected empty station ID on error, got %q", station.Id)
	}
}
