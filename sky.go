package main

import (
	"log"
	"net/http"
	"strconv"
	"time"
	"slices"

	"github.com/ugjka/go-tz/v2"
	"github.com/Code-Hex/synchro/iso8601"
)

type queryHandler struct{}

func (qh queryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	queryParams := r.URL.Query()
	
	t, tErr := iso8601.ParseDateTime(queryParams.Get("time"))

	lon, err := strconv.ParseFloat(queryParams.Get("lon"), 64)
	if err != nil && tErr != nil {
		http.Error(w, "Invalid longitude", http.StatusBadRequest)
		return
	}

	lat, err := strconv.ParseFloat(queryParams.Get("lat"), 64)
	if err != nil && tErr != nil {
		http.Error(w, "Invalid latitude", http.StatusBadRequest)
		return
	}

	if tErr != nil {
		t, tErr = calTime(lon, lat)
		if tErr != nil {
			http.Error(w, "Invalid query", http.StatusBadRequest)
		}
	}

	weather := queryParams.Get("weather")
	weatherTypes := []string{"clear", "sunny", "cloudy", "rain", "snow", "fog", "thunder"}
	if result := slices.Contains(weatherTypes, weather); result == false {
		weather = "unknown"
	}

	w.Write([]byte(t.Format(time.RFC3339)))
	w.Write([]byte(weather))
}

func calTime(longitude float64, latitude float64) (time.Time, error) {
	zone, err := tz.GetZone(tz.Point{
		Lon: longitude, Lat: latitude,
	})
	if err != nil {
		return time.Time{}, err
	}

	location, err := time.LoadLocation(zone[0])
	if err != nil {
		return time.Time{}, err
	}
	return time.Now().In(location), nil
}

func main() {
	s := http.Server{
		Addr:         ":8080",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 90 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      queryHandler{},
	}
	log.Println("Listening Port", s.Addr)
	err := s.ListenAndServe()
	if err != nil {
		if err != http.ErrServerClosed {
			panic(err)
		}
	}
}
