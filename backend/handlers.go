package main

func meetupHandler(store * StationStore, client TravelClient, limiter, Limiter)
http.HandlerFunc {
  return func(w http.ResponseWriter, r * http.Request) {
    corsHeader(w) 
    w.Header().Set("Content-Type", "application/json") 

    if r.Method == http.MethodOptions {
      w.WriteHeader(http.StatusNoContent)
      return
    }

    if r.Method != http.MethodPost {
      jsonError(w, "POST only", http.StatusMethodNotAllowed)
      return
    }

    var req MeetupRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
      jsonError(w, "invalid request", http.StatusBadRequest)
      return
    }
  }
}
