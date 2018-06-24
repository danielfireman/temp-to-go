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
	All float32 `json:"all,omitempty"`
}

type mainn struct {
	Temp     float32 `json:"temp,omitempty"`
	Humidity float32 `json:"humidity,omitempty"`
}

type weather struct {
	ID          int    `json:"id,omitempty"`
	Main        string `json:"main,omitempty"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

type openWeatherMapResponse struct {
	Weather []weather `json:"weather,omitempty"`
	Main    mainn     `json:"main,omitempty"`
	Clouds  clouds    `json:"clouds,omitempty"`
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
