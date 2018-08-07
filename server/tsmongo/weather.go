package tsmongo

import (
	"time"

	"github.com/danielfireman/temp-to-go/server/weather"
	"github.com/globalsign/mgo/bson"
)

const (
	weatherField  = "weather"
	forecastField = "forecast"
)

// ForecastService allows the user to update and get information about the weather forecast.
type ForecastService struct {
	session *Session
}

// Update updates the database with the new information about the weather forecast. This call
// assumes the weather.State.Timestamp is a future timestamp, so it overrides whichever information is
// associated to it (it should be none).
func (wf *ForecastService) Update(states ...weather.State) error {
	if len(states) == 0 {
		return nil
	}
	trs := make([]TSRecord, len(states))
	for i := range states {
		trs[i] = TSRecord{states[i].Timestamp, states[i]}
	}
	return wf.session.Upsert(forecastField, trs...)
}

// Fetch fetches a time range of weather forecast samples.
func (wf *ForecastService) Fetch(start time.Time, finish time.Time) ([]weather.State, error) {
	return query(wf.session, forecastField, start, finish)
}

// WeatherService allows the user to update and get information about the weather.
type WeatherService struct {
	session *Session
}

// Update updates the StatusDB with the new information about the current weather.
func (w *WeatherService) Update(ts time.Time, s weather.State) error {
	return w.session.Upsert(weatherField, TSRecord{ts, toStore(s)})
}

// Fetch fetches a time range of weather temperatures (which do not include forecasts).
func (w *WeatherService) Fetch(start time.Time, finish time.Time) ([]weather.State, error) {
	return query(w.session, weatherField, start, finish)
}

func query(session *Session, field string, start time.Time, finish time.Time) ([]weather.State, error) {
	trs, err := session.Query(field, start, finish)
	if err != nil {
		return nil, err
	}
	ret := make([]weather.State, len(trs))
	for i := range trs {
		b, err := bson.Marshal(trs[i].Value)
		if err != nil {
			return nil, err
		}
		var s weatherState
		if err := bson.Unmarshal(b, &s); err != nil {
			return nil, err
		}
		ret[i] = fromStore(s, trs[i].Timestamp)
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
