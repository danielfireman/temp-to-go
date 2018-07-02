package db

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const scheduledInfoDBCollectionName = "sidb"

// ScheduledInfoDB stores the result of the collection of information that happens at a pre-determined
// schedule. For instance, fetching the current weather information and the bedroom temperature.
type ScheduledInfoDB struct {
	session    *mgo.Session
	collection *mgo.Collection
}

// NewCurrentWeather updates the ScheduledInfoDB with the new information about the current weather.
func (db *ScheduledInfoDB) NewCurrentWeather(cw CurrentWeather) error {
	// Inspiration: https://www.mongodb.com/blog/post/schema-design-for-time-series-data-in-mongodb
	now := time.Now().In(time.UTC)
	th := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
	_, err := db.collection.Upsert(
		bson.M{"timestamp_hour": th},
		bson.M{
			"timestamp_hour":  th,
			"current_weather": cw,
		},
	)
	return err
}

// Close terminates the ScheduledInfoDB session. It's a runtime error to use a session
// after it has been closed.
func (db *ScheduledInfoDB) Close() {
	db.session.Close()
}

// DialScheduledInfoDB sets up a connection to the database specified by the passed-in URI.
func DialScheduledInfoDB(uri string) (*ScheduledInfoDB, error) {
	info, err := mgo.ParseURL(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid db URI:\"%s\" err:%q", uri, err)
	}
	s, err := mgo.DialWithInfo(info)
	if err != nil {
		return nil, err
	}
	s.SetMode(mgo.Monotonic, true)
	return &ScheduledInfoDB{
		session:    s,
		collection: s.DB(info.Database).C(scheduledInfoDBCollectionName),
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

// CurrentWeather stores the complete information about the weather at a certain time.
type CurrentWeather struct {
	Description WeatherDescription `bson:"description,omitempty"`
	Wind        Wind               `bson:"wind,omitempty"`
	Temp        float32            `bson:"temp,omitempty"`       // Temperature, Celsius
	Humidity    float32            `bson:"humidity,omitempty"`   // Humidity, %
	Rain        float32            `bson:"rain,omitempty"`       // Rain volume for the last hours
	Cloudiness  float32            `bson:"cloudiness,omitempty"` // Cloudiness, %
}
