package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

type RateLimiter struct {
	mu    sync.Mutex
	count int
	limit int
	file  string
	month int
	year  int
}

type rateLimiterState struct {
	Count int `json:"count"`
	Month int `json:"month"`
	Year  int `json:"year"`
}

func newRateLimiter(limit int, file string) *RateLimiter {
	rl := &RateLimiter{limit: limit, file: file}
	rl.load()
	return rl
}

func (rl *RateLimiter) load() {
	now := time.Now()

	data, err := os.ReadFile(rl.file)
	if err != nil {
		rl.count = 0
		rl.month = int(now.Month())
		rl.year = now.Year()
		rl.save()
		return
	}

	var state rateLimiterState
	if err := json.Unmarshal(data, &state); err != nil {
		rl.count = 0
		rl.month = int(now.Month())
		rl.year = now.Year()
		return
	}

	if state.Year < now.Year() || state.Month < int(now.Month()) {
		log.Printf("New month detected — resetting search count")
		rl.count = 0
		rl.month = int(now.Month())
		rl.year = now.Year()
		rl.save()
		return
	}

	rl.count = state.Count
	rl.month = state.Month
	rl.year = state.Year
	log.Printf("Rate limiter loaded — %d searches used, %d remaining", rl.count, rl.limit-rl.count)
}

func (rl *RateLimiter) save() {
	state := rateLimiterState{
		Count: rl.count,
		Month: rl.month,
		Year:  rl.year,
	}
	data, _ := json.Marshal(state)
	os.WriteFile(rl.file, data, 0644)
}

func (rl *RateLimiter) allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	if rl.year < now.Year() || rl.month < int(now.Month()) {
		log.Printf("New month detected mid-session — resetting search count")
		rl.count = 0
		rl.month = int(now.Month())
		rl.year = now.Year()
		rl.save()
	}

	if rl.count >= rl.limit {
		return false
	}
	rl.count++
	rl.save()
	return true
}

func (rl *RateLimiter) remaining() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.limit - rl.count
}
