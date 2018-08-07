package main

import (
	"log"
	"os"
	"time"

	"github.com/danielfireman/temp-to-go/server/tsmongo"
	"github.com/sajari/regression"
)

func main() {
	mgoURI := os.Getenv("MONGODB_URI")
	if mgoURI == "" {
		log.Fatalf("Invalid MONGODB_URI: %s", mgoURI)
	}
	session, err := tsmongo.Dial(mgoURI)
	if err != nil {
		log.Fatalf("Error connecting to status DB: %s", mgoURI)
	}
	log.Println("Connected to timeseries mongo.")
	defer session.Close()

	bedroomService := tsmongo.NewBedroomService(session)
	weatherService := tsmongo.NewWeatherService(session)
	fanService := tsmongo.NewFanService(session)
	forecastService := tsmongo.NewForecastService(session)
	predictionService := tsmongo.NewPredictionService(session)

	// Get the last week worth of data.
	st := time.Now().Add(-7 * 24 * time.Hour)
	et := time.Now()

	r := newRegression()
	r.Train(getTrainSet(st, et, weatherService, fanService, bedroomService)...)

	// Run model.
	r.Run()

	// Predict the next 24 hours and updates the database.
	predictions := predict(r, et, forecastService)
	if err := predictionService.Update(predictions...); err != nil {
		log.Fatalf("Error updating predictions:%q", err)
	}
}

func newRegression() regression.Regression {
	var r regression.Regression
	r.SetObserved("Bedroom temperature")
	r.SetVar(0, "Weather Temperature")
	r.SetVar(1, "Fan State")
	return r
}

func predInputs(ws, fs float64) []float64 {
	return []float64{ws, fs}
}

func predOrDie(r regression.Regression, w, f float64) float64 {
	p, err := r.Predict(predInputs(w, f))
	if err != nil {
		log.Fatalf("Error predicting temperature with fan:%.0f weather:%f :%q", f, w, err)
	}
	return p
}

func predict(r regression.Regression, et time.Time, forecastService *tsmongo.ForecastService) []tsmongo.Prediction {
	forecast, err := forecastService.Fetch(et, et.Add(24*time.Hour))
	if err != nil {
		log.Fatal(err)
	}
	var predictions []tsmongo.Prediction
	for _, f := range forecast {
		predictions = append(predictions, tsmongo.Prediction{
			TempFanOff:  predOrDie(r, f.Temp, 0),
			TempFanLow:  predOrDie(r, f.Temp, 1),
			TempFanHigh: predOrDie(r, f.Temp, 2),
		})
	}
	return predictions
}

func getTrainSet(st, et time.Time, weatherService *tsmongo.WeatherService, fanService *tsmongo.FanService, bedroomService *tsmongo.BedroomService) regression.DataPoints {
	bs, err := bedroomService.FetchState(st, et)
	if err != nil {
		log.Fatalf("Error fetching past bedroom temperature: %q", err)
	}
	if len(bs) == 0 {
		log.Fatalf("Can not predict without any bedroom temperature.")
	}
	ws, err := weatherService.Fetch(st, et)
	if err != nil {
		log.Fatalf("Error fetching past weather: %q", err)
	}
	fs, err := fanService.FetchState(st, et)
	if err != nil {
		log.Fatalf("Error fetching fan state: %q", err)
	}
	wsMap := make(map[time.Time]float64)
	for _, s := range ws {
		wsMap[s.Timestamp] = s.Temp
	}
	fsMap := fillFanState(fs, st, et)
	var trainSet regression.DataPoints
	for _, b := range bs {
		// Only consider valid for the train set there is a full tuple.
		t := b.Timestamp
		f, fok := fsMap[t]
		w, wok := wsMap[t]
		if fok && wok {
			trainSet = append(trainSet, regression.DataPoint(b.Temperature, predInputs(w, f)))
		}
	}
	return trainSet
}

func fillFanState(fs []tsmongo.FanState, st, et time.Time) map[time.Time]float64 {
	m := make(map[time.Time]float64)
	// First populate the map with what we have.
	for _, s := range fs {
		m[s.Timestamp] = fanStatusToFloat(s.Status)
	}
	// Fill the map, always using the last known status.
	lastStatus := float64(0)
	if len(fs) > 0 {
		lastStatus = fanStatusToFloat(fs[0].Status)
	}
	for currTime := st.In(time.UTC); currTime.Before(et); currTime = currTime.Add(1 * time.Hour) {
		if val, ok := m[currTime]; ok {
			lastStatus = val
		} else {
			m[currTime] = lastStatus
		}
	}
	return m
}

func fanStatusToFloat(s tsmongo.FanStatus) float64 {
	switch s {
	case tsmongo.FanLowSpeed:
		return 1
	case tsmongo.FanHighSpeed:
		return 2
	default:
		return 0
	}
}
