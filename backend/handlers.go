package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"sync"
)

func meetupHandler(store *StationStore, apiKey string, limiter *RedisLimiter) http.HandlerFunc {
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

		stationA, okA := store.Find(req.StationA)
		stationB, okB := store.Find(req.StationB)
		if !okA || !okB {
			json.NewEncoder(w).Encode(MeetupResponse{Error: "station not found"})
			return
		}

		if stationA.Name == stationB.Name {
			json.NewEncoder(w).Encode(MeetupResponse{Error: "please choose two different stations"})
			return
		}

		allStations := store.All()
		originA := fmt.Sprintf("%.6f,%.6f", stationA.Lat, stationA.Lng)
		originB := fmt.Sprintf("%.6f,%.6f", stationB.Lat, stationB.Lng)

		var timesA, timesB []int
		var errA, errB error

		if os.Getenv("USE_MOCK") == "true" {
			log.Printf("MOCK: returning fake travel times for %s and %s", stationA.Name, stationB.Name)
			timesA = mockTravelTimes(allStations)
			timesB = mockTravelTimes(allStations)
		} else {
			if !limiter.allow() {
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(MeetupResponse{Error: "monthly search limit reached"})
				return
			}

			log.Printf("Real API call - %d searches remaining", limiter.remaining())

			var wg sync.WaitGroup
			wg.Add(2)

			go func() {
				defer wg.Done()
				timesA, errA = fetchTravelTimes(apiKey, originA, allStations)
			}()

			go func() {
				defer wg.Done()
				timesB, errB = fetchTravelTimes(apiKey, originB, allStations)
			}()

			wg.Wait()

			if errA != nil || errB != nil {
				json.NewEncoder(w).Encode(MeetupResponse{Error: "failed to fetch travel times"})
				return
			}
		}

		var candidates []CandidateResult
		const (
			minTimeWeight = 0.5
			diffWeight    = 0.3
			maxTimeWeight = 0.2
		)
		for i, s := range allStations {
			if s.Name == stationA.Name || s.Name == stationB.Name {
				continue
			}

			tA := timesA[i]
			tB := timesB[i]
			if tA >= 99999 || tB >= 99999 {
				continue
			}

			diff := int(math.Abs(float64(tA - tB)))
			minTime := tA
			if tB < minTime {
				minTime = tB
			}

			maxTime := tA
			if tB > maxTime {
				maxTime = tB
			}

			score := float64(minTime)*minTimeWeight + float64(diff)*diffWeight + float64(maxTime)*maxTimeWeight
			candidates = append(candidates, CandidateResult{
				Station:    s,
				TimeFromA:  tA,
				TimeFromB:  tB,
				MinTime:    minTime,
				MaxTime:    maxTime,
				Difference: diff,
				Score:      score,
			})
		}

		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].Score < candidates[j].Score
		})

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

func stationsHandler(store *StationStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(store.All())
	}
}
