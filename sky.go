package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/ugjka/go-tz/v2"
)

type HelloHandler struct{}

func (hh HelloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	queryParams := r.URL.Query()
	lon, err := strconv.ParseFloat(queryParams.Get("longitude"), 64)
	if err != nil {
		http.Error(w, "Invalid longitude", http.StatusBadRequest)
		return
	}
	lat, err := strconv.ParseFloat(queryParams.Get("latitude"), 64)
	if err != nil {
		http.Error(w, "Invalid latitude", http.StatusBadRequest)
		return
	}

	t, err := calTime(lon, lat)
	if(err != nil) {
		http.Error(w, "Invalid TimeZone", http.StatusBadRequest)
	}
	w.Write([]byte(t.Format(time.RFC3339)))
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
		Handler:      HelloHandler{},
	}
	log.Println("Listening Port", s.Addr)
	err := s.ListenAndServe()
	if err != nil {
		if err != http.ErrServerClosed {
			panic(err)
		}
	}
}
