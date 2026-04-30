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
						"name": "New York",
						"lat": 40.7128,
						"lon": -74.0060,
						"country": "US",
						"state": "NY"
					}
				]`)),
			}, nil
		},
	}

	locations, err := comms.GetLocationsWithClient(mockHTTPClient, "New York", "test-key")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(locations) == 0 {
		t.Error("expected locations, got empty response")
	}
	if len(locations) > 0 && locations[0].Name != "New York" {
		t.Errorf("expected name 'New York', got %q", locations[0].Name)
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

	locations, err := comms.GetLocationsWithClient(mockHTTPClient, "New York", "invalid-key")
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

	locations, err := comms.GetLocationsWithClient(mockHTTPClient, "New York", "test-key")
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

	locations, err := comms.GetLocationsWithClient(mockHTTPClient, "New York", "test-key")
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
		Name: "New York",
		Lat:  40.7128,
		Lon:  -74.0060,
	}

	station, err := comms.GetStationWithClient(mockHTTPClient, location, "test-token")
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
		Name: "New York",
		Lat:  40.7128,
		Lon:  -74.0060,
	}

	station, err := comms.GetStationWithClient(mockHTTPClient, location, "test-token")
	if err == nil {
		t.Error("expected error for server error, got nil")
	}
	if station.Id != "" {
		t.Errorf("expected empty station ID on error, got %q", station.Id)
	}
}
