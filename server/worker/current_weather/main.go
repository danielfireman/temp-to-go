package main

import (
	"log"
	"os"
	"time"

	"github.com/danielfireman/temp-to-go/server/status"
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
	sdb, err := status.DialDB(mgoURI)
	if err != nil {
		log.Fatalf("Error connecting to status DB: %s", mgoURI)
	}
	defer sdb.Close()
	log.Println("Connected to StatusDB.")

	ws, err := weatherClient.Current()
	if err != nil {
		log.Fatalf("Error retrieving current weather: %q", err)
	}
	if err := sdb.StoreWeather(time.Now(), ws); err != nil {
		log.Fatalf("Error updating status with current weather: %q", err)
	}
	log.Printf("Succefully updated status with current weather: %+v\n", ws)
}
