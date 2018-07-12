package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"

	"github.com/danielfireman/temp-to-go/server/db"
)

func main() {
	owmKey := os.Getenv("OWM_API_KEY")
	if owmKey == "" {
		log.Fatalf("Invalid OWM_API_KEY: %s", owmKey)
	}
	mgoURI := os.Getenv("MONGODB_URI")
	if mgoURI == "" {
		log.Fatalf("Invalid MONGODB_URI: %s", mgoURI)
	}
	sdb, err := db.DialStatusDB(mgoURI)
	if err != nil {
		log.Fatalf("Error connecting to ScheduledInfoDB: %s", mgoURI)
	}
	defer sdb.Close()
	log.Println("Connected to StatusDB.")
	resp, err := http.Get(fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?units=metric&q=Maceio,BR&appid=%s", owmKey))
	if err != nil {
		log.Fatalf("Error fetching current weather from OWM: %q", err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading OMW response body: %q", err)
	}
	var omwResp openWeatherMapResponse
	if err := json.Unmarshal(b, &omwResp); err != nil {
		log.Fatalf("Error unmarshalling OMW response: %q", err)
	}
	log.Printf("Succefully feched response from OWM: %+v\n", omwResp)
	ws := db.WeatherStatus{
		Description: db.WeatherDescription{
			Text: omwResp.Weather[0].Description,
			Icon: omwResp.Weather[0].Icon,
		},
		Wind: db.Wind{
			Speed:     round(omwResp.Wind.Speed),
			Direction: round(omwResp.Wind.Deg),
		},
		Temp:       round(omwResp.Main.Temp),
		Humidity:   round(omwResp.Main.Humidity),
		Rain:       round(omwResp.Rain.ThreeHours),
		Cloudiness: round(omwResp.Clouds.All),
	}
	if err := sdb.StoreWeatherStatus(ws); err != nil {
		log.Fatalf("Error updating ScheduledInfoDB: %q", err)
	}
	log.Printf("Succefully updated ScheduledInfoDB: %+v\n", ws)
}

type clouds struct {
	//  Cloudiness, %
	All float64 `json:"all,omitempty"`
}

type mainn struct {
	Temp     float64 `json:"temp,omitempty"`     // Temperature, Celsius
	Humidity float64 `json:"humidity,omitempty"` // Humidity, %
}

type wind struct {
	Speed float64 `json:"speed,omitempty"` // Wind speed, meter/sec
	Deg   float64 `json:"deg,omitempty"`   // Wind direction, degrees (meteorological)
}

type rain struct {
	ThreeHours float64 `json:"3h,omitempty"` // Rain volume for the last 3 hours
}

type weather struct {
	Description string `json:"description,omitempty"` // Weather condition within the group
	Icon        string `json:"icon,omitempty"`        // Weather icon id
}

type openWeatherMapResponse struct {
	Weather []weather `json:"weather,omitempty"`
	Main    mainn     `json:"main,omitempty"`
	Clouds  clouds    `json:"clouds,omitempty"`
	Wind    wind      `json:"wind,omitempty"`
	Rain    rain      `json:"rain,omitempty"`
	Dt      int64     `json:"dt,omitempty"` // Time of data calculation, unix, UTC
}

// round receives a float64, rounds it to 4 most signigicant digits (max) and returns it as
// float32. Mostly used to decrease massive (unnecessary) precision of float64 and thus
// to decrease storage requirements.
func round(f float64) float32 {
	return float32(math.Round(f/0.0001) * 0.0001)
}
