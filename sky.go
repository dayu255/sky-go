package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"slices"
	"strconv"
	"time"
	"encoding/json"

	"github.com/Code-Hex/synchro/iso8601"
	"github.com/dayu255/sun-level"
	"github.com/ugjka/go-tz/v2"
)

type skyColor struct {
	Color1 string
	Color2 string
	Color3 string
	Color4 string
}

// type skyContext struct {
// 	scene     string
// 	weather   string
// 	cloudNess float32
// 	sunLevel  int
// }

type skyData struct {
	Color        skyColor
	CssGradient string
// 	context      skyContext
}

type responseData struct {
	Success bool
	Data skyData
}

func clamp(v int) int {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}

func curve(x, a, b, c float64) int {
	return clamp(int(a*((x-b)*(x-b)) + c))
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

func makeColor(x float64, p *[4][3][3]float64) skyColor {
	result := [4]string{}
	for i := 0; i < 4; i++ {
		r := curve(x, p[i][0][0], p[i][0][1], p[i][0][2])
		g := curve(x, p[i][1][0], p[i][1][1], p[i][1][2])
		b := curve(x, p[i][2][0], p[i][2][1], p[i][2][2])
		result[i] = fmt.Sprintf("#%02x%02x%02x", r, g, b)
	}
	return skyColor{
		Color1: result[0],
		Color2: result[1],
		Color3: result[2],
		Color4: result[3],
	}
}

func NewSky(t time.Time, weather string, sunLevel float64) skyColor {
	sunLevel = sunLevel / math.Pi * 180
	s := [4][3][3]float64{
		{{0, 0, 0}, {0, 0, 0}, {0, 2, 40}},
		{{0, 0, 0}, {0, 0, 0}, {-0.0190476, 3.80952, 40}},
		{{0, 0, 0}, {0, 0, 0}, {0, 4, 40}},
		{{-0.0666667, 0.333333, 30}, {}, {0.133333, 5.33333, 40}},
	}
	return makeColor(sunLevel, &s)
}

type queryHandler struct{}

func (qh queryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r)
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
	d := responseData{
		Success: true,
		Data: skyData{
			Color: sky,
			CssGradient: fmt.Sprintf("linear-gradient(to bottom, %s, %s, %s, %s)", sky.Color1, sky.Color2, sky.Color3, sky.Color4),
		},
	}	


	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", "*")
	jsonBytes, err :=json.MarshalIndent(d, "", "	")
	if err != nil {
		w.Write([]byte("Json encoding error."))
		return
	}

	w.Write(jsonBytes)
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
