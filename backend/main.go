package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// ─── Data Types ───────────────────────────────────────────────────────────────

type Station struct {
	Name  string   `json:"name"`
	Code  string   `json:"code"`
	Lines []string `json:"lines"`
	Lat   float64  `json:"lat"`
	Lng   float64  `json:"lng"`
}

type MeetupRequest struct {
	StationA string `json:"stationA"`
	StationB string `json:"stationB"`
}

type CandidateResult struct {
	Station    Station `json:"station"`
	TimeFromA  int     `json:"timeFromA"`
	TimeFromB  int     `json:"timeFromB"`
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

// ─── Load Stations ────────────────────────────────────────────────────────────

func loadStations() ([]Station, error) {
	data, err := os.ReadFile("stations.json")
	if err != nil {
		return nil, err
	}
	var stations []Station
	err = json.Unmarshal(data, &stations)
	return stations, err
}

func findStation(stations []Station, name string) (Station, bool) {
	for _, s := range stations {
		if s.Name == name {
			return s, true
		}
	}
	return Station{}, false
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

// ─── Handlers ─────────────────────────────────────────────────────────────────

func meetupHandler(stations []Station, apiKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.Method != http.MethodPost {
			json.NewEncoder(w).Encode(MeetupResponse{Error: "POST only"})
			return
		}

		var req MeetupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			json.NewEncoder(w).Encode(MeetupResponse{Error: "invalid request"})
			return
		}

		stationA, okA := findStation(stations, req.StationA)
		stationB, okB := findStation(stations, req.StationB)

		if !okA || !okB {
			json.NewEncoder(w).Encode(MeetupResponse{Error: "station not found"})
			return
		}

		if stationA.Name == stationB.Name {
			json.NewEncoder(w).Encode(MeetupResponse{Error: "please choose two different stations"})
			return
		}

		originA := fmt.Sprintf("%.6f,%.6f", stationA.Lat, stationA.Lng)
		originB := fmt.Sprintf("%.6f,%.6f", stationB.Lat, stationB.Lng)

		// Fetch travel times from both stations in parallel
		var timesA, timesB []int
		var errA, errB error

		if os.Getenv("USE_MOCK") == "true" {
			// Mock mode — no API calls, no cost
			log.Printf("MOCK: returning fake travel times for %s and %s", stationA.Name, stationB.Name)
			timesA = mockTravelTimes(stations)
			timesB = mockTravelTimes(stations)
		} else {
			// Real mode — calls Google Distance Matrix API
			var wg sync.WaitGroup
			wg.Add(2)

			go func() {
				defer wg.Done()
				timesA, errA = fetchTravelTimes(apiKey, originA, stations)
			}()

			go func() {
				defer wg.Done()
				timesB, errB = fetchTravelTimes(apiKey, originB, stations)
			}()

			wg.Wait()

			if errA != nil || errB != nil {
				json.NewEncoder(w).Encode(MeetupResponse{Error: "failed to fetch travel times"})
				return
			}
		}

		// Score every candidate station
		var candidates []CandidateResult
		for i, s := range stations {
			if s.Name == stationA.Name || s.Name == stationB.Name {
				continue
			}

			tA := timesA[i]
			tB := timesB[i]
			if tA >= 99999 || tB >= 99999 {
				continue
			}

			diff := int(math.Abs(float64(tA - tB)))
			maxTime := tA
			if tB > maxTime {
				maxTime = tB
			}

			score := float64(diff)*0.6 + float64(maxTime)*0.4

			candidates = append(candidates, CandidateResult{
				Station:    s,
				TimeFromA:  tA,
				TimeFromB:  tB,
				MaxTime:    maxTime,
				Difference: diff,
				Score:      score,
			})
		}

		// Sort by score, lowest is best
		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].Score < candidates[j].Score
		})

		// Return top 5
		if len(candidates) > 5 {
			candidates = candidates[:5]
		}

		json.NewEncoder(w).Encode(MeetupResponse{
			StationA:   stationA,
			StationB:   stationB,
			Candidates: candidates,
		})
	}
}

func stationsHandler(stations []Station) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stations)
	}
}

// ─── Main ─────────────────────────────────────────────────────────────────────

func main() {
	godotenv.Load()

	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_MAPS_API_KEY is not set")
	}
	log.Printf("API key loaded ✓")

	stations, err := loadStations()
	if err != nil {
		log.Fatalf("Failed to load stations: %v", err)
	}
	log.Printf("Loaded %d stations ✓", len(stations))

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/api/stations", stationsHandler(stations))
	mux.HandleFunc("/api/find-meetup", meetupHandler(stations, apiKey))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Backend running on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
