package main

import (
	"log"
	"os"
	"time"

	"github.com/danielfireman/temp-to-go/server/status"
	"github.com/sajari/regression"
)

func main() {
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

	// Get the last week worth of data.
	st := time.Now().Add(-7 * 24 * time.Hour)
	et := time.Now()

	var r regression.Regression
	r.SetObserved("Bedroom temperature")
	r.SetVar(0, "Weather Temperature")
	r.SetVar(1, "Fan State")
	r.SetVar(2, "Timestamp")

	// ### Trainning.
	bs, err := sdb.Bedroom().FetchState(st, et)
	if err != nil {
		log.Fatalf("Error fetching past bedroom temperature: %q", err)
	}
	if len(bs) == 0 {
		log.Fatalf("Can not predict without any bedroom temperature.")
	}

	// Normalize the start and end timestamp (to match the bedroom ones)
	st, et = bs[0].Timestamp, bs[len(bs)-1].Timestamp
	ws, err := sdb.Weather().Fetch(st, et)
	if err != nil {
		log.Fatalf("Error fetching past weather: %q", err)
	}
	wsMap := make(map[time.Time]float64)
	for _, s := range ws {
		wsMap[s.Timestamp] = s.Temp
	}
	fs, err := sdb.Fan().FetchState(st, et)
	if err != nil {
		log.Fatalf("Error fetching fan state: %q", err)
	}
	fsMap := fillFanState(fs, st, et)
	var trainSet regression.DataPoints
	for _, b := range bs {
		t := b.Timestamp
		f, fok := fsMap[t]
		w, wok := wsMap[t]
		if fok && wok {
			trainSet = append(trainSet, regression.DataPoint(b.Temperature, []float64{w, f, float64(b.Timestamp.Unix())}))
		}
	}
	r.Train(trainSet...)

	// Run model.
	r.Run()

	// Predict the next 24 hours and updates the database.
	forecast, err := sdb.WeatherForecast().Fetch(et, et.Add(24*time.Hour))
	if err != nil {
		log.Fatal(err)
	}
	var predictions []status.Prediction
	for _, f := range forecast {
		predictions = append(predictions, status.Prediction{
			TempFanOff:  predOrDie(r, f.Temp, 0),
			TempFanLow:  predOrDie(r, f.Temp, 1),
			TempFanHigh: predOrDie(r, f.Temp, 2),
		})
	}
	if err := sdb.Predictions().Update(predictions...); err != nil {
		log.Fatalf("Error updating predictions:%q", err)
	}
}

func predOrDie(r regression.Regression, temp, fanMode float64) float64 {
	p, err := r.Predict([]float64{temp, fanMode})
	if err != nil {
		log.Fatalf("Error predicting temperature with fan %.0f:%q", fanMode, err)
	}
	return p
}

func fillFanState(fs []status.FanState, st, et time.Time) map[time.Time]float64 {
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

func fanStatusToFloat(s status.FanStatus) float64 {
	switch s {
	case status.FanLowSpeed:
		return 1
	case status.FanHighSpeed:
		return 2
	default:
		return 0
	}
}
