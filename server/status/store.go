package status

import (
	"fmt"
	"time"

	"github.com/danielfireman/temp-to-go/server/weather"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	statusDBCollectionName = "sdb"
	timestampIndexField    = "timestamp_hour"
)

const (
	bedroomField  = "bedroom"
	weatherField  = "weather"
	forecastField = "forecast"
	fanField      = "fan"
)

// DB stores the result of the collection of information that happens at a pre-determined
// schedule. For instance, fetching the current weather information and the bedroom temperature.
type DB struct {
	session    *mgo.Session
	collection *mgo.Collection
}

// StoreWeather updates the StatusDB with the new information about the current weather.
func (db *DB) StoreWeather(ts time.Time, s weather.State) error {
	return db.store(ts, weatherField, toStore(s))
}

// FetchWeather fetches a time range of weather temperatures (which do not include forecasts).
func (db *DB) FetchWeather(start time.Time, finish time.Time) ([]weather.State, error) {
	iter := db.iterRange(start, finish, weatherField)
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

// StoreWeatherForecast updates the StatusDB with the new information about the weather forecast. This call
// assumes the weather.State.Timestamp is a future timestamp, so it overrides whichever information is
// associated to it (it should be none).
func (db *DB) StoreWeatherForecast(states ...weather.State) error {
	bulk := db.collection.Bulk()
	for _, s := range states {
		db.bulkStore(bulk, s.Timestamp, forecastField, toStore(s))
	}
	_, err := bulk.Run()
	return err
}

// StoreBedroomTemperature updates the StatusDB with the new bedroom temperature.
func (db *DB) StoreBedroomTemperature(ts time.Time, temp float32) error {
	return db.store(time.Now(), bedroomField, temp)
}

func (db *DB) iterRange(start time.Time, finish time.Time, typez string) *mgo.Iter {
	startUTC := start.In(time.UTC)
	finishUTC := finish.In(time.UTC)
	return db.collection.Find(
		bson.M{
			timestampIndexField: bson.M{
				"$gte": startUTC,
				"$lte": finishUTC,
			},
			"type": typez,
		}).Iter()
}

// Fan returns a Fan instance.
func (db *DB) Fan() *Fan {
	return &Fan{db}
}

func hourUTC(ts time.Time) time.Time {
	tUTC := ts.In(time.UTC)
	return time.Date(tUTC.Year(), tUTC.Month(), tUTC.Day(), tUTC.Hour(), 0, 0, 0, tUTC.Location())
}

func (db *DB) store(ts time.Time, field string, val interface{}) error {
	// Inspiration: https://www.mongodb.com/blog/post/schema-design-for-time-series-data-in-mongodb
	utc := hourUTC(ts)
	_, err := db.collection.Upsert(
		bson.M{timestampIndexField: utc, "type": field},
		bson.M{
			timestampIndexField: utc,
			"type":              field,
			"value":             val,
		},
	)
	return err
}

func (db *DB) bulkStore(bulk *mgo.Bulk, ts time.Time, field string, val interface{}) {
	utc := hourUTC(ts)
	bulk.Upsert(
		bson.M{timestampIndexField: utc, "type": field},
		bson.M{
			timestampIndexField: utc,
			"type":              field,
			"value":             val,
		},
	)
}

// Close terminates the ScheduledInfoDB session. It's a runtime error to use a session
// after it has been closed.
func (db *DB) Close() {
	db.session.Close()
}

// DialDB sets up a connection to the database specified by the passed-in URI.
func DialDB(uri string) (*DB, error) {
	info, err := mgo.ParseURL(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid db URI:\"%s\" err:%q", uri, err)
	}
	s, err := mgo.DialWithInfo(info)
	if err != nil {
		return nil, err
	}
	s.SetMode(mgo.Monotonic, true)
	return &DB{
		session:    s,
		collection: s.DB(info.Database).C(statusDBCollectionName),
	}, nil
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

type bedroomTemp float64

type bedroomTempDocument struct {
	Value bedroomTemp `bson:"value,omitempty"`
}

type weatherDocument struct {
	Timestamp time.Time    `bson:"timestamp_hour,omitempty"`
	Value     weatherState `bson:"value,omitempty"`
}
