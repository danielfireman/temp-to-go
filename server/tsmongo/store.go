package tsmongo

import (
	"fmt"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	statusDBCollectionName = "sdb"
	timestampIndexField    = "timestamp_hour"
	typeField              = "type"
	valueField             = "value"
)

// Session represents a connection to a mongo timeseries collection.
type Session struct {
	session *mgo.Session
	dbName  string
	col     *mgo.Collection
}

// NewSession creates a new Session instance, which allows the communication to the underlying
// timeseries mongo database;
func NewSession(s *mgo.Session, dbName string) *Session {
	return &Session{s.Copy(), dbName, s.DB(dbName).C(statusDBCollectionName)}
}

// TSRecord represents a value to be added to timeseries database.
type TSRecord struct {
	Timestamp time.Time   `bson:"timestamp_hour,omitempty"`
	Value     interface{} `bson:"value,omitempty"`
}

// Dial sets up a connection to the specified timeseries database specified by the passed-in URI.
func Dial(uri string) (*Session, error) {
	info, err := mgo.ParseURL(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid db URI:\"%s\" err:%q", uri, err)
	}
	s, err := mgo.DialWithInfo(info)
	if err != nil {
		return nil, err
	}
	s.SetMode(mgo.Monotonic, true)
	dbName := info.Database
	return NewSession(s, dbName), nil
}

// Upsert inserts the given data into the timeseries database overriding the data if necessary.
// If more than one values are passed, it performs a bulk-upsert.
func (s *Session) Upsert(field string, val ...TSRecord) error {
	// Inspiration: https://www.mongodb.com/blog/post/schema-design-for-time-series-data-in-mongodb
	switch len(val) {
	case 0:
		return nil
	case 1:
		utc := hourUTC(val[0].Timestamp)
		_, err := s.col.Upsert(
			bson.M{timestampIndexField: utc, typeField: field},
			bson.M{
				timestampIndexField: utc,
				typeField:           field,
				valueField:          val[0].Value,
			},
		)
		return err
	default:
		bulk := s.col.Bulk()
		for _, v := range val {
			utc := hourUTC(v.Timestamp)
			bulk.Upsert(
				bson.M{timestampIndexField: utc, typeField: field},
				bson.M{
					timestampIndexField: utc,
					typeField:           field,
					valueField:          v.Value,
				},
			)
		}
		_, err := bulk.Run()
		return err
	}
}

// Query fetches all records from timeseries mongo with the specified range.
func (s *Session) Query(field string, start time.Time, finish time.Time) ([]TSRecord, error) {
	startUTC := start.In(time.UTC)
	finishUTC := finish.In(time.UTC)
	iter := s.col.Find(
		bson.M{
			timestampIndexField: bson.M{
				"$gte": startUTC,
				"$lte": finishUTC,
			},
			typeField: field,
		}).Sort("-" + timestampIndexField).Iter()

	var ret []TSRecord
	var d TSRecord
	for iter.Next(&d) {
		ret = append(ret, d)
	}
	if err := iter.Close(); err != nil {
		if err == mgo.ErrNotFound {
			return ret, nil
		}
		return nil, fmt.Errorf("Error querying tsmongo within range(%v,%v): %q", start, finish, err)
	}
	return ret, nil
}

// Last returns the last element in the timeseries, if any.
func (s *Session) Last(field string) (TSRecord, error) {
	var r TSRecord
	err := s.col.Find(bson.M{typeField: field}).Sort("-" + timestampIndexField).One(&r)
	if err != nil {
		return TSRecord{}, err
	}
	return r, nil
}

// Close release resources associated with this connection.
func (s *Session) Close() {
	s.session.Close()
}

// Copy works just like New, but preserves the database and any authentication
// information from the original session.
func (s *Session) Copy() *Session {
	c := s.session.Copy()
	dbName := s.dbName
	return NewSession(c, dbName)
}

func hourUTC(ts time.Time) time.Time {
	tUTC := ts.In(time.UTC)
	return time.Date(tUTC.Year(), tUTC.Month(), tUTC.Day(), tUTC.Hour(), 0, 0, 0, tUTC.Location())
}

// NewBedroomService creates a new BedroomService, which allows to interact with the bedroom field
// of the timeseries.
func NewBedroomService(s *Session) *BedroomService {
	return &BedroomService{s}
}

// NewFanService creates a new BedroomService, which allows to interact with the fan field
// of the timeseries.
func NewFanService(s *Session) *FanService {
	return &FanService{s}
}

// NewWeatherService creates a new WeatherService, which allows to interact with the weather field
// of the timeseries.
func NewWeatherService(s *Session) *WeatherService {
	return &WeatherService{s}
}

// NewForecastService creates a new ForecastService, which allows to interact with the forecast field
// of the timeseries.
func NewForecastService(s *Session) *ForecastService {
	return &ForecastService{s}
}

// NewPredictionService creates a new PredictionService, which allows to interact with the predictions field
// of the timeseries.
func NewPredictionService(s *Session) *PredictionService {
	return &PredictionService{s}
}
