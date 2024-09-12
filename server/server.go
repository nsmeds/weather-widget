package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nsmeds/weather-widget/comms"
)

type Server struct {
	*http.Server
	apiKey string
	// TODO logger
	// TODO metrics
}

func New(host string, port int, apiKey string) *Server {
	s := Server{
		apiKey: apiKey,
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
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "internal system error"}`))
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
		body, err := io.ReadAll(r.Body) // TODO some kind of sanitization
		if err != nil {
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "internal system error"}`))
			return
		}
		res, err := comms.GetLocation(string(body), s.apiKey)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fmt.Printf("res: %#v", res)
		locationResponse, err := json.Marshal(res)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(locationResponse))
	}
}


