package main

import (
	"encoding/json"
	"os"
)

type Station struct {
	Name  string   `json:"name"`
	Code  string   `json:"code"`
	Lines []string `json:"lines"`
	Lat   float64  `json:"lat"`
	Lng   float64  `json:"lng"`
}

type StationStore struct {
	list  []Station
	index map[string]Station
}

func NewStationStore(path string) (*StationStore, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var stations []Station
	if err := json.Unmarshal(data, &stations); err != nil {
		return nil, err
	}

	index := make(map[string]Station, len(stations))
	for _, station := range stations {
		index[station.Name] = station
	}

	return &StationStore{list: stations, index: index}, nil
}

func (s *StationStore) Find(name string) (Station, bool) {
	station, ok := s.index[name]
	return station, ok
}

func (s *StationStore) Candidates(a, b Station) []Station {
	candidates := make([]Station, 0, len(s.list))
	for _, station := range s.list {
		if station.Name == a.Name || station.Name == b.Name {
			continue
		}
		candidates = append(candidates, station)
	}
	return candidates
}

func (s *StationStore) All() []Station {
	stations := make([]Station, len(s.list))
	copy(stations, s.list)
	return stations
}
