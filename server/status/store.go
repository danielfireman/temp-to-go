package status

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const statusDBCollectionName = "sdb"

// DB stores the result of the collection of information that happens at a pre-determined
// schedule. For instance, fetching the current weather information and the bedroom temperature.
type DB struct {
	session    *mgo.Session
	collection *mgo.Collection
}

// StoreWeather updates the StatusDB with the new information about the current weather.
func (db *DB) StoreWeather(ws WeatherStatus) error {
	return db.store(hourUTC(time.Now()), "weather", ws)
}

// StoreBedroomTemperature updates the StatusDB with the new bedroom temperature.
func (db *DB) StoreBedroomTemperature(temp float32) error {
	now := hourUTC(time.Now())
	var s status
	if err := db.collection.Find(bson.M{"timestamp_hour": now}).One(&s); err != nil {
		return err
	}
	s.Bedroom.Temp = temp
	return db.store(now, "bedroom", s.Bedroom)
}

func hourUTC(t time.Time) time.Time {
	tUTC := t.In(time.UTC)
	return time.Date(tUTC.Year(), tUTC.Month(), tUTC.Day(), tUTC.Hour(), 0, 0, 0, tUTC.Location())
}

func (db *DB) store(t time.Time, field string, val interface{}) error {
	// Inspiration: https://www.mongodb.com/blog/post/schema-design-for-time-series-data-in-mongodb
	_, err := db.collection.Upsert(
		bson.M{"timestamp_hour": t},
		bson.M{
			"timestamp_hour": t,
			field:            val,
		},
	)
	return err
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

// Wind stores information about the wind.
type Wind struct {
	Speed     float32 `bson:"speed,omitempty"` // Wind speed, meter/sec
	Direction float32 `bson:"deg,omitempty"`   // Wind direction, degrees (meteorological)
}

// WeatherDescription stores overall information to describe the weather. That include text, images and so on.
type WeatherDescription struct {
	Text string `bson:"text,omitempty"` // Weather condition within the group
	Icon string `bson:"icon,omitempty"` // Weather icon id
}

// WeatherStatus stores the complete information about the weather at a certain time.
type WeatherStatus struct {
	Description WeatherDescription `bson:"description,omitempty"`
	Wind        Wind               `bson:"wind,omitempty"`
	Temp        float32            `bson:"temp,omitempty"`       // Temperature, Celsius
	Humidity    float32            `bson:"humidity,omitempty"`   // Humidity, %
	Rain        float32            `bson:"rain,omitempty"`       // Rain volume for the last hours
	Cloudiness  float32            `bson:"cloudiness,omitempty"` // Cloudiness, %
}

// BedroomStatus stores information about the bedroom.
type BedroomStatus struct {
	Temp float32 `bson:"temp,omitempty"` // Temperature, Celsius
	Fan  byte    `bson:"fan,omitempty"`  // On, Off (1, 0)
}

type status struct {
	Bedroom BedroomStatus `bson:"bedroom,omitempty"`
	Weather WeatherStatus `bson:"weather,omitempty"`
}
