package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nsmeds/weather-widget/comms"
)

type Server struct {
	*http.Server
	commsClient   *comms.CommsClient
	geoCodeApiKey string
	weatherApiKey string
	// TODO logger
	// TODO metrics
}

// WeatherResponse is the unified response returned by the /weather endpoint.
type WeatherResponse struct {
	Location locationData         `json:"location"`
	Current  comms.CurrentConditions `json:"current"`
	Forecast comms.TodayForecast     `json:"forecast"`
	Averages comms.ClimateNormals    `json:"averages"`
}

type locationData struct {
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
}

func New(host string, port int, geocodeApiKey, weatherApiKey string) *Server {
	s := Server{
		commsClient:   comms.NewClientWithDefaults(),
		geoCodeApiKey: geocodeApiKey,
		weatherApiKey: weatherApiKey,
	}
	httpServer := http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: s.Routes(),
	}
	s.Server = &httpServer
	return &s
}

func (s *Server) handleDefaultRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		processStartedAt := time.Now().Format(time.RFC3339Nano)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "internal system error"})
			return
		}
		message := fmt.Sprintf("received %v at %s", string(body), processStartedAt)
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(message))
	}
}

func (s *Server) handleWeatherRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "internal system error"})
			return
		}
		query := strings.TrimSpace(string(body))
		if query == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "location query required"})
			return
		}

		locations, err := s.commsClient.GetLocations(query, s.geoCodeApiKey)
		if err != nil {
			fmt.Println(err)
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "location lookup failed"})
			return
		}
		if len(locations) == 0 {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": "location not found"})
			return
		}
		if len(locations) > 1 {
			writeJSON(w, http.StatusOK, locations)
			return
		}

		loc := locations[0]

		// Fetch CDO station ID for climate normals (best-effort; proceed if unavailable)
		station, err := s.commsClient.GetStation(loc, s.weatherApiKey)
		if err != nil {
			fmt.Println("GetStation:", err)
		}

		// Fetch NWS grid point metadata to get forecast and observation station URLs
		points, err := s.commsClient.GetNWSPoints(loc.Lat, loc.Lon)
		if err != nil {
			fmt.Println(err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "failed to retrieve weather grid data"})
			return
		}

		nwsStationID, err := s.commsClient.GetNWSObservationStation(points.ObservationStationsURL)
		if err != nil {
			fmt.Println(err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "failed to find observation station"})
			return
		}

		current, err := s.commsClient.GetCurrentObservation(nwsStationID)
		if err != nil {
			fmt.Println(err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "failed to retrieve current conditions"})
			return
		}

		forecast, err := s.commsClient.GetNWSForecast(points.ForecastURL)
		if err != nil {
			fmt.Println(err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "failed to retrieve forecast"})
			return
		}

		// Climate normals are best-effort; an empty StationInfo produces zero values
		var normals comms.ClimateNormals
		if station.Id != "" {
			normals, err = s.commsClient.GetClimateNormals(station.Id, time.Now(), s.weatherApiKey)
			if err != nil {
				fmt.Println("GetClimateNormals:", err)
			}
		}

		name := loc.Name
		if loc.State != "" {
			name = loc.Name + ", " + loc.State
		}
		writeJSON(w, http.StatusOK, WeatherResponse{
			Location: locationData{Name: name, Lat: loc.Lat, Lon: loc.Lon},
			Current:  current,
			Forecast: forecast,
			Averages: normals,
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		fmt.Println("writeJSON encode error:", err)
	}
}
