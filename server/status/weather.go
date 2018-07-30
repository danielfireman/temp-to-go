package status

import (
	"fmt"
	"time"

	"github.com/danielfireman/temp-to-go/server/weather"
	"github.com/globalsign/mgo"
)

const (
	weatherField  = "weather"
	forecastField = "forecast"
)

// WeatherForecast allows the user to update and get information about the weather forecast.
type WeatherForecast struct {
	db *DB
}

// Update updates the database with the new information about the weather forecast. This call
// assumes the weather.State.Timestamp is a future timestamp, so it overrides whichever information is
// associated to it (it should be none).
func (wf *WeatherForecast) Update(states ...weather.State) error {
	bulk := wf.db.collection.Bulk()
	for _, s := range states {
		wf.db.bulkStore(bulk, s.Timestamp, forecastField, toStore(s))
	}
	_, err := bulk.Run()
	return err
}

// Fetch fetches a time range of weather forecast samples.
func (wf *WeatherForecast) Fetch(start time.Time, finish time.Time) ([]weather.State, error) {
	iter := wf.db.iterRange(start, finish, forecastField)
	var ret []weather.State
	var d weatherDocument
	for iter.Next(&d) {
		ret = append(ret, fromStore(d.Value, d.Timestamp))
	}
	if err := iter.Close(); err != nil {
		if err == mgo.ErrNotFound {
			return ret, nil
		}
		return nil, fmt.Errorf("Error fetching weather forecast (%v, %v): %q", start, finish, err)
	}
	return ret, nil
}

// Weather allows the user to update and get information about the weather.
type Weather struct {
	db *DB
}

// Update updates the StatusDB with the new information about the current weather.
func (w *Weather) Update(ts time.Time, s weather.State) error {
	return w.db.store(ts, weatherField, toStore(s))
}

// Fetch fetches a time range of weather temperatures (which do not include forecasts).
func (w *Weather) Fetch(start time.Time, finish time.Time) ([]weather.State, error) {
	iter := w.db.iterRange(start, finish, weatherField)
	var ret []weather.State
	var d weatherDocument
	for iter.Next(&d) {
		ret = append(ret, fromStore(d.Value, d.Timestamp))
	}
	if err := iter.Close(); err != nil {
		if err == mgo.ErrNotFound {
			return ret, nil
		}
		return nil, fmt.Errorf("Error fetching weather: %q", err)
	}
	return ret, nil
}

func toStore(s weather.State) weatherState {
	return weatherState{
		Description: weatherDescription{
			Text: s.Description.Text,
			Icon: s.Description.Icon,
		},
		Wind: wind{
			Speed:     s.Wind.Speed,
			Direction: s.Wind.Direction,
		},
		Temp:       s.Temp,
		Humidity:   s.Humidity,
		Rain:       s.Rain,
		Cloudiness: s.Cloudiness,
	}
}

func fromStore(s weatherState, hour time.Time) weather.State {
	return weather.State{
		Description: weather.Description{
			Text: s.Description.Text,
			Icon: s.Description.Icon,
		},
		Wind: weather.Wind{
			Speed:     s.Wind.Speed,
			Direction: s.Wind.Direction,
		},
		Temp:       s.Temp,
		Humidity:   s.Humidity,
		Rain:       s.Rain,
		Cloudiness: s.Cloudiness,
		Timestamp:  hour,
	}
}

// weatherState stores the complete information about the weather at a certain time.
type weatherState struct {
	Description weatherDescription `bson:"description,omitempty"`
	Wind        wind               `bson:"wind,omitempty"`
	Temp        float64            `bson:"temp,omitempty"`       // Temperature, Celsius
	Humidity    float64            `bson:"humidity,omitempty"`   // Humidity, %
	Rain        float64            `bson:"rain,omitempty"`       // Rain volume for the last hours
	Cloudiness  float64            `bson:"cloudiness,omitempty"` // Cloudiness, %
}

type wind struct {
	Speed     float64 `bson:"speed,omitempty"` // Wind speed, meter/sec
	Direction float64 `bson:"deg,omitempty"`   // Wind direction, degrees (meteorological)
}

type weatherDescription struct {
	Text string `bson:"text,omitempty"` // Weather condition within the group
	Icon string `bson:"icon,omitempty"` // Weather icon id
}

type weatherDocument struct {
	Timestamp time.Time    `bson:"timestamp_hour,omitempty"`
	Value     weatherState `bson:"value,omitempty"`
}
