package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// ─── Data Types ───────────────────────────────────────────────────────────────

type MeetupRequest struct {
	StationA string `json:"stationA"`
	StationB string `json:"stationB"`
}

type CandidateResult struct {
	Station    Station `json:"station"`
	TimeFromA  int     `json:"timeFromA"`
	TimeFromB  int     `json:"timeFromB"`
	MinTime    int     `json:"minTime"`
	MaxTime    int     `json:"maxTime"`
	Difference int     `json:"difference"`
	Score      float64 `json:"score"`
}

type MeetupResponse struct {
	StationA   Station           `json:"stationA"`
	StationB   Station           `json:"stationB"`
	Candidates []CandidateResult `json:"candidates"`
	Error      string            `json:"error,omitempty"`
}

// ─── Google Distance Matrix ───────────────────────────────────────────────────

type GMapsResponse struct {
	Rows []struct {
		Elements []struct {
			Status   string `json:"status"`
			Duration struct {
				Value int `json:"value"`
			} `json:"duration"`
		} `json:"elements"`
	} `json:"rows"`
	Status string `json:"status"`
}

func fetchTravelTimes(apiKey, origin string, destinations []Station) ([]int, error) {
	times := make([]int, len(destinations))
	client := &http.Client{Timeout: 10 * time.Second}

	// Google allows max 25 destinations per request so we batch them
	batchSize := 25
	for start := 0; start < len(destinations); start += batchSize {
		end := start + batchSize
		if end > len(destinations) {
			end = len(destinations)
		}
		batch := destinations[start:end]

		// Build destination string
		destStr := ""
		for i, d := range batch {
			if i > 0 {
				destStr += "|"
			}
			destStr += fmt.Sprintf("%.6f,%.6f", d.Lat, d.Lng)
		}

		params := url.Values{}
		params.Set("origins", origin)
		params.Set("destinations", destStr)
		params.Set("mode", "transit")
		params.Set("key", apiKey)

		apiURL := "https://maps.googleapis.com/maps/api/distancematrix/json?" + params.Encode()

		resp, err := client.Get(apiURL)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var gmaps GMapsResponse
		if err := json.NewDecoder(resp.Body).Decode(&gmaps); err != nil {
			return nil, err
		}

		if gmaps.Status != "OK" {
			return nil, fmt.Errorf("Google API error: %s", gmaps.Status)
		}

		for i, el := range gmaps.Rows[0].Elements {
			if el.Status == "OK" {
				times[start+i] = el.Duration.Value
			} else {
				times[start+i] = 99999
			}
		}
	}

	return times, nil
}

func mockTravelTimes(destinations []Station) []int {
	times := make([]int, len(destinations))
	for i := range times {
		// Random travel time between 5 and 60 minutes in seconds
		times[i] = (rand.Intn(55) + 5) * 60
	}
	return times
}

// ─── Main ─────────────────────────────────────────────────────────────────────

func main() {
	godotenv.Load()

	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_MAPS_API_KEY is not set")
	}
	log.Printf("API key loaded ✓")

	store, err := NewStationStore("stations.json")
	if err != nil {
		log.Fatalf("Failed to load stations: %v", err)
	}
	log.Printf("Loaded %d stations ✓", len(store.All()))

	limiter := newRedisLimiter(30)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/api/remaining", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"remaining": limiter.remaining()})
	})
	mux.HandleFunc("/api/stations", stationsHandler(store))
	mux.HandleFunc("/api/find-meetup", meetupHandler(store, apiKey, limiter))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Backend running on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
