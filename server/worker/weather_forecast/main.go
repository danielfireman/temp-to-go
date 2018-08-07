package main

import (
	"log"
	"os"

	"github.com/danielfireman/temp-to-go/server/tsmongo"
	"github.com/danielfireman/temp-to-go/server/weather"
)

func main() {
	owmKey := os.Getenv("OWM_API_KEY")
	if owmKey == "" {
		log.Fatalf("Invalid OWM_API_KEY: %s", owmKey)
	}
	weatherClient := weather.NewOWMClient(owmKey)

	mgoURI := os.Getenv("MONGODB_URI")
	if mgoURI == "" {
		log.Fatalf("Invalid MONGODB_URI: %s", mgoURI)
	}
	session, err := tsmongo.Dial(mgoURI)
	if err != nil {
		log.Fatalf("Error connecting to status DB: %s", mgoURI)
	}
	defer session.Close()
	forecastService := tsmongo.NewForecastService(session)
	log.Println("Connected to StatusDB.")

	ws, err := weatherClient.Forecast()
	if err != nil {
		log.Fatalf("Error retrieving weather forecast weather: %q", err)
	}
	if err := forecastService.Update(ws...); err != nil {
		log.Fatalf("Error updating status with weather forecast: %q", err)
	}
	log.Printf("Succefully updated status with weather forecast: %+v\n", ws)
}
