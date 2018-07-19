package status

import (
	"fmt"
	"time"

	"github.com/danielfireman/temp-to-go/server/weather"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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

// StoreWeatherForecast updates the StatusDB with the new information about the weather forecast. This call
// assumes the weather.State.Timestamp is a future timestamp, so it overrides whichever information is
// associated to it (it should be none).
func (db *DB) StoreWeatherForecast(states ...weather.State) error {
	bulk := db.collection.Bulk()
	for _, s := range states {
		ts := time.Unix(s.Timestamp, 0)
		db.bulkStore(bulk, ts, forecastField, toStore(s))
	}
	_, err := bulk.Run()
	return err
}

// StoreBedroomTemperature updates the StatusDB with the new bedroom temperature.
func (db *DB) StoreBedroomTemperature(ts time.Time, temp float32) error {
	return db.store(time.Now(), bedroomField, temp)
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

// weatherState stores the complete information about the weather at a certain time.
type weatherState struct {
	Description weatherDescription `bson:"description,omitempty"`
	Wind        wind               `bson:"wind,omitempty"`
	Temp        float32            `bson:"temp,omitempty"`       // Temperature, Celsius
	Humidity    float32            `bson:"humidity,omitempty"`   // Humidity, %
	Rain        float32            `bson:"rain,omitempty"`       // Rain volume for the last hours
	Cloudiness  float32            `bson:"cloudiness,omitempty"` // Cloudiness, %
}

type wind struct {
	Speed     float32 `bson:"speed,omitempty"` // Wind speed, meter/sec
	Direction float32 `bson:"deg,omitempty"`   // Wind direction, degrees (meteorological)
}

type weatherDescription struct {
	Text string `bson:"text,omitempty"` // Weather condition within the group
	Icon string `bson:"icon,omitempty"` // Weather icon id
}
