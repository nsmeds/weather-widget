package server_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nsmeds/weather-widget/server"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name          string
		host          string
		port          int
		geocodeAPIKey string
		weatherAPIKey string
	}{
		{
			name:          "create server with valid params",
			host:          "localhost",
			port:          8080,
			geocodeAPIKey: "test-key-1",
			weatherAPIKey: "test-key-2",
		},
		{
			name:          "create server with different host",
			host:          "0.0.0.0",
			port:          9000,
			geocodeAPIKey: "key1",
			weatherAPIKey: "key2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := server.New(tt.host, tt.port, tt.geocodeAPIKey, tt.weatherAPIKey)
			if srv == nil {
				t.Fatal("expected server, got nil")
			}
			if srv.Addr != fmt.Sprintf("%s:%d", tt.host, tt.port) {
				t.Errorf("expected addr %s:%d, got %s", tt.host, tt.port, srv.Addr)
			}
		})
	}
}

func TestHandleDefaultRequest(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
		expectedHeader string
	}{
		{
			name:           "POST request with body",
			method:         http.MethodPost,
			body:           `{"query": "test"}`,
			expectedStatus: http.StatusOK,
			expectedHeader: "application/json",
		},
		{
			name:           "GET request",
			method:         http.MethodGet,
			body:           "",
			expectedStatus: http.StatusOK,
			expectedHeader: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := server.New("localhost", 8080, "key1", "key2")
			handler := srv.Routes().ServeHTTP

			var reqBody io.Reader
			if tt.body != "" {
				reqBody = bytes.NewBufferString(tt.body)
			}

			req := httptest.NewRequest(tt.method, "/", reqBody)
			w := httptest.NewRecorder()

			handler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			if contentType := w.Header().Get("content-type"); contentType != tt.expectedHeader {
				t.Errorf("expected content-type %s, got %s", tt.expectedHeader, contentType)
			}
			if w.Body.Len() == 0 {
				t.Error("expected response body, got empty")
			}
		})
	}
}

func TestHandleWeatherRequest(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{
			name:           "weather request with body",
			body:           "test location",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "weather request with empty body",
			body:           "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := server.New("localhost", 8080, "", "")
			handler := srv.Routes().ServeHTTP

			var reqBody io.Reader
			if tt.body != "" {
				reqBody = bytes.NewBufferString(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/weather", reqBody)
			w := httptest.NewRecorder()

			handler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestRoutes(t *testing.T) {
	t.Run("routes returns a valid mux", func(t *testing.T) {
		srv := server.New("localhost", 8080, "key1", "key2")
		mux := srv.Routes()
		if mux == nil {
			t.Fatal("expected mux, got nil")
		}
	})

	t.Run("root route exists", func(t *testing.T) {
		srv := server.New("localhost", 8080, "key1", "key2")
		mux := srv.Routes()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code == http.StatusNotFound {
			t.Error("expected root route to be handled, got 404")
		}
	})

	t.Run("weather route exists", func(t *testing.T) {
		srv := server.New("localhost", 8080, "", "")
		mux := srv.Routes()

		req := httptest.NewRequest(http.MethodPost, "/weather", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code == http.StatusNotFound {
			t.Error("expected /weather route to be handled, got 404")
		}
	})
}
