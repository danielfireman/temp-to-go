package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type clouds struct {
	//  Cloudiness, %
	All float32 `json:"all,omitempty"`
}

type mainn struct {
	Temp     float32 `json:"temp,omitempty"`     // Temperature, Celsius
	Humidity float32 `json:"humidity,omitempty"` // Humidity, %
}

type wind struct {
	Speed     float32 `json:"speed,omitempty"` // Wind speed, meter/sec
	Direction float32 `json:"deg,omitempty"`   // Wind direction, degrees (meteorological)
}

type rain struct {
	ThreeHours float32 `json:"3h,omitempty"` // Rain volume for the last 3 hours
}

type weather struct {
	ID          int    `json:"id,omitempty"`          // Weather condition id
	Main        string `json:"main,omitempty"`        // Group of weather parameters (Rain, Snow, Extreme etc.)
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

func main() {
	owmKey := os.Getenv("OWM_API_KEY")
	if owmKey == "" {
		log.Fatalf("Invalid OWM_API_KEY: %s", owmKey)
	}
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
	fmt.Printf("%+v", omwResp)
}
