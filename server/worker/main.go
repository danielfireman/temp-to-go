package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
		log.Fatalf("Invalid MONGODB_URI: %s", owmKey)
	}
	scheddb, err := db.DialScheduledInfoDB(mgoURI)
	if err != nil {
		log.Fatalf("Error connecting to ScheduledInfoDB: %s", owmKey)
	}
	defer scheddb.Close()
	log.Println("Connected to ScheduledInfoDB.")
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
	cw := db.CurrentWeather{
		Description: db.WeatherDescription{
			Description: omwResp.Weather[0].Description,
			Icon:        omwResp.Weather[0].Icon,
		},
		Wind: db.Wind{
			Speed:     omwResp.Wind.Speed,
			Direction: omwResp.Wind.Deg,
		},
		Temp:       omwResp.Main.Temp,
		Humidity:   omwResp.Main.Humidity,
		Rain:       omwResp.Rain.ThreeHours,
		Cloudiness: omwResp.Clouds.All,
	}
	if err := scheddb.NewCurrentWeather(cw); err != nil {
		log.Fatalf("Error updating ScheduledInfoDB: %q", err)
	}
	log.Printf("Succefully updated ScheduledInfoDB: %+v\n", cw)
}

type clouds struct {
	//  Cloudiness, %
	All float32 `json:"all,omitempty"`
}

type mainn struct {
	Temp     float32 `json:"temp,omitempty"`     // Temperature, Celsius
	Humidity float32 `json:"humidity,omitempty"` // Humidity, %
}

type wind struct {
	Speed float32 `json:"speed,omitempty"` // Wind speed, meter/sec
	Deg   float32 `json:"deg,omitempty"`   // Wind direction, degrees (meteorological)
}

type rain struct {
	ThreeHours float32 `json:"3h,omitempty"` // Rain volume for the last 3 hours
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
