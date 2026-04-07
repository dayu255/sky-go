package main

import (
	"log"
	"net/http"
	"strconv"
	"time"
	"slices"

	"github.com/dayu255/sun-level"
	"github.com/ugjka/go-tz/v2"
	"github.com/Code-Hex/synchro/iso8601"
)

type color struct{
	color1 string
	color2 string
	color3 string
	color4 string
}

type context struct {
	scene string
	weather string
	cloudNess float32
	sunLevel int
}

type data struct {
	color color
	css_gradient string
	context context
}


func decimalToHex(n int) string {
	if n == 0 {
		return "0"
	}

	chars := "0123456789ABCDEF"

	var ans string
	for n > 0 {
		ans = string(chars[n % 16]) + ans
		n /= 16
	}
	
	return ans
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

func NewSky(t time.Time, weather string, sunLevel float64) color {
			
}

type queryHandler struct{}

func (qh queryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	queryParams := r.URL.Query()
	
	t, tErr := iso8601.ParseDateTime(queryParams.Get("time"))
	if tErr != nil {
		t = time.Now()
	}

	lon, lonErr := strconv.ParseFloat(queryParams.Get("lon"), 64)
	lat, latErr := strconv.ParseFloat(queryParams.Get("lat"), 64)

	weather := queryParams.Get("weather")
	weatherTypes := []string{"clear", "sunny", "cloudy", "rain", "snow", "fog", "thunder"}
	if result := slices.Contains(weatherTypes, weather); result == false {
		weather = "unknown"
	}

	if tErr != nil && lonErr != nil && latErr != nil {
		w.Write([]byte("Query Error. We need latitude and longitude or time."))
		return
	}

	sky := NewSky(t, weather, sun.CalSunLevel(t, lat, lon))
	
	w.Write()
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
