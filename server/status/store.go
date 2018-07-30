package status

import (
	"fmt"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	statusDBCollectionName = "sdb"
	timestampIndexField    = "timestamp_hour"
)

// DB stores the result of the collection of information that happens at a pre-determined
// schedule. For instance, fetching the current weather information and the bedroom temperature.
type DB struct {
	session    *mgo.Session
	collection *mgo.Collection
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

// Fan returns a Fan instance, which allows to update and fetch information about the fan state.
func (db *DB) Fan() *Fan {
	return &Fan{db}
}

// Bedroom returns a Bedroom instance, which allows to update and fetch information about the bedroom state.
func (db *DB) Bedroom() *Bedroom {
	return &Bedroom{db}
}

// WeatherForecast returns a WeatherForecast instance, which allows to update and fetch information about the weather forecast.
func (db *DB) WeatherForecast() *WeatherForecast {
	return &WeatherForecast{db}
}

// Weather returns a WeatherForecast instance, which allows to update and fetch information about weather.
func (db *DB) Weather() *Weather {
	return &Weather{db}
}

// Predictions returns a Prediction instance, which allows to update and fetch information about predictions.
func (db *DB) Predictions() *Predictions {
	return &Predictions{db}
}

// Close terminates the ScheduledInfoDB session. It's a runtime error to use a session
// after it has been closed.
func (db *DB) Close() {
	db.session.Close()
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
		}).Sort("-" + timestampIndexField).Iter()
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
