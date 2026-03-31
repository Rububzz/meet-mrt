package main

type Station struct {
  Name string `json:"name"`
  Code string `json:"code"`
  Lines []string `json:"lines"`
  Lat float64 `json:"lat"`
  Lng float64 `json:"lng"`
}

type StationStore struct {
  list []Station
  index map[string]Station
}

func NewStationStore(path string) (*StationStore, error) {
  data, err := os.ReadFile(path)
}

func (s *StationStore) Find(name string) (Station, bool) {

}

func (*s StationStore) Candidates(a, b Station) []Station {

}


