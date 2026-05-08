package comms_test

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/nsmeds/weather-widget/comms"
)

func TestGetClimateNormalsSuccess(t *testing.T) {
	mockHTTPClient := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(strings.NewReader(`{
					"results": [
						{"datatype": "DLY-TMAX-NORMAL", "value": 669},
						{"datatype": "DLY-TMIN-NORMAL", "value": 493}
					]
				}`)),
			}, nil
		},
	}

	client := comms.NewClient(mockHTTPClient)
	date := time.Date(2026, time.May, 7, 0, 0, 0, 0, time.UTC)
	normals, err := client.GetClimateNormals("GHCND:USW00014739", date, "test-token")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// CDO NORMAL_DLY values are in tenths of °F
	if normals.TemperatureHighF != 66.9 {
		t.Errorf("expected high 66.9°F, got %v", normals.TemperatureHighF)
	}
	if normals.TemperatureLowF != 49.3 {
		t.Errorf("expected low 49.3°F, got %v", normals.TemperatureLowF)
	}
	if normals.Period == "" {
		t.Error("expected non-empty period label")
	}
}

func TestGetClimateNormalsAPIError(t *testing.T) {
	mockHTTPClient := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       io.NopCloser(strings.NewReader(`{"message": "unauthorized"}`)),
			}, nil
		},
	}

	client := comms.NewClient(mockHTTPClient)
	date := time.Date(2026, time.May, 7, 0, 0, 0, 0, time.UTC)
	_, err := client.GetClimateNormals("GHCND:USW00014739", date, "bad-token")
	if err == nil {
		t.Error("expected error for unauthorized response, got nil")
	}
}

func TestGetClimateNormalsNetworkError(t *testing.T) {
	mockHTTPClient := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return nil, io.EOF
		},
	}

	client := comms.NewClient(mockHTTPClient)
	date := time.Date(2026, time.May, 7, 0, 0, 0, 0, time.UTC)
	_, err := client.GetClimateNormals("GHCND:USW00014739", date, "test-token")
	if err == nil {
		t.Error("expected error for network failure, got nil")
	}
}

func TestGetClimateNormalsNoResults(t *testing.T) {
	mockHTTPClient := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"results": []}`)),
			}, nil
		},
	}

	client := comms.NewClient(mockHTTPClient)
	date := time.Date(2026, time.May, 7, 0, 0, 0, 0, time.UTC)
	_, err := client.GetClimateNormals("GHCND:USW00014739", date, "test-token")
	if err == nil {
		t.Error("expected error when no results returned, got nil")
	}
}

func TestGetClimateNormalsUsesCorrectQueryDate(t *testing.T) {
	var capturedURL string
	mockHTTPClient := &mockClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			capturedURL = req.URL.String()
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(strings.NewReader(`{
					"results": [
						{"datatype": "DLY-TMAX-NORMAL", "value": 500},
						{"datatype": "DLY-TMIN-NORMAL", "value": 300}
					]
				}`)),
			}, nil
		},
	}

	client := comms.NewClient(mockHTTPClient)
	// Use a specific date to verify the query uses the normals base year
	date := time.Date(2026, time.March, 15, 0, 0, 0, 0, time.UTC)
	_, err := client.GetClimateNormals("GHCND:USW00014739", date, "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should use month/day from provided date with normals base year
	if !strings.Contains(capturedURL, "03-15") {
		t.Errorf("expected query to contain '03-15', got URL: %s", capturedURL)
	}
}
