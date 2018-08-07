package main

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/danielfireman/temp-to-go/server/tsmongo"
	"github.com/danielfireman/temp-to-go/server/weather"
	"github.com/globalsign/mgo/dbtest"
)

var mongoDB dbtest.DBServer

func TestPredictor(t *testing.T) {
	tempDir, _ := ioutil.TempDir("", "predictor_testing")
	defer func() { os.RemoveAll(tempDir) }()

	mongoDB.SetPath(tempDir)
	defer mongoDB.Stop()
	defer mongoDB.Wipe()

	s := mongoDB.Session()
	defer s.Close()
	dbNames, _ := s.DatabaseNames()
	dbName := dbNames[0]

	session := tsmongo.NewSession(s, dbName)
	defer session.Close()

	startTime := time.Now()
	endTime := startTime.Add(12 * time.Hour)

	bs := tsmongo.NewBedroomService(session)
	bs.UpdateTemperature(startTime.Add(1*time.Hour), 27)
	bs.UpdateTemperature(startTime.Add(2*time.Hour), 28)
	bs.UpdateTemperature(startTime.Add(3*time.Hour), 25) // User switches the fan on (high).
	bs.UpdateTemperature(startTime.Add(4*time.Hour), 25)
	bs.UpdateTemperature(startTime.Add(5*time.Hour), 27) // User switches the fan off.
	bs.UpdateTemperature(startTime.Add(6*time.Hour), 28.9)
	bs.UpdateTemperature(startTime.Add(7*time.Hour), 30)
	bs.UpdateTemperature(startTime.Add(8*time.Hour), 31)
	bs.UpdateTemperature(startTime.Add(9*time.Hour), 27) // User switches the fan on (low).
	bs.UpdateTemperature(startTime.Add(10*time.Hour), 26.5)
	bs.UpdateTemperature(startTime.Add(11*time.Hour), 26)
	bs.UpdateTemperature(startTime.Add(12*time.Hour), 27) // User switches the fan off.

	fs := tsmongo.NewFanService(session)
	fs.UpdateStatus(startTime.Add(3*time.Hour), tsmongo.FanHighSpeed)
	fs.UpdateStatus(startTime.Add(5*time.Hour), tsmongo.FanOff)
	fs.UpdateStatus(startTime.Add(9*time.Hour), tsmongo.FanLowSpeed)
	fs.UpdateStatus(startTime.Add(12*time.Hour), tsmongo.FanOff)

	// The temperature outside changes very little during the considered period.
	ws := tsmongo.NewWeatherService(session)
	ws.Update(weather.State{Timestamp: startTime.Add(1 * time.Hour), Temp: 27})
	ws.Update(weather.State{Timestamp: startTime.Add(2 * time.Hour), Temp: 27.5})
	ws.Update(weather.State{Timestamp: startTime.Add(3 * time.Hour), Temp: 28})
	ws.Update(weather.State{Timestamp: startTime.Add(4 * time.Hour), Temp: 29})
	ws.Update(weather.State{Timestamp: startTime.Add(5 * time.Hour), Temp: 28.8})
	ws.Update(weather.State{Timestamp: startTime.Add(6 * time.Hour), Temp: 28.6})
	ws.Update(weather.State{Timestamp: startTime.Add(7 * time.Hour), Temp: 29})
	ws.Update(weather.State{Timestamp: startTime.Add(8 * time.Hour), Temp: 30})
	ws.Update(weather.State{Timestamp: startTime.Add(9 * time.Hour), Temp: 29})
	ws.Update(weather.State{Timestamp: startTime.Add(10 * time.Hour), Temp: 28})
	ws.Update(weather.State{Timestamp: startTime.Add(11 * time.Hour), Temp: 27})
	ws.Update(weather.State{Timestamp: startTime.Add(12 * time.Hour), Temp: 27.5})

	// The current forecast source (OWM) only provides foreacasts for every 3 hours.
	// We are simulating like the evening forecast, with temperatures dropping.
	forecasts := []weather.State{
		weather.State{Timestamp: endTime.Add(3 * time.Hour), Temp: 26},
		weather.State{Timestamp: endTime.Add(6 * time.Hour), Temp: 25},
		weather.State{Timestamp: endTime.Add(9 * time.Hour), Temp: 24},
		weather.State{Timestamp: endTime.Add(12 * time.Hour), Temp: 23},
		weather.State{Timestamp: endTime.Add(15 * time.Hour), Temp: 22},
	}
	fcs := tsmongo.NewForecastService(session)
	fcs.Update(forecasts...)
	r := newRegression()
	r.Train(getTrainSet(startTime, endTime, ws, fs, bs)...)
	r.Run()
	predictions := predict(r, endTime, fcs)
	if len(predictions) != len(forecasts) {
		t.Errorf("len(predictions) want:%d got:%d", len(forecasts), len(predictions))
	}
	for _, p := range predictions {
		if p.TempFanOff < p.TempFanLow {
			t.Errorf("p.TempFanOff > p.TempFanLow p.TempFanOff:%f p.TempFanLow:%f", p.TempFanOff, p.TempFanLow)
		}
		if p.TempFanLow < p.TempFanHigh {
			t.Errorf("p.TempFanOff > p.TempFanLow p.TempFanLow:%f p.TempFanHifh:%f", p.TempFanLow, p.TempFanHigh)
		}
	}
}
