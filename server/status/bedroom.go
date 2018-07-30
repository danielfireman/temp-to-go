package status

import (
	"fmt"
	"time"

	"github.com/globalsign/mgo"
)

const bedroomField = "bedroom"

// BedroomState state stores the bedroom state (e.g. temperature, humidity) at a certain moment.
type BedroomState struct {
	Temperature float64   `json:"temp,omitempty"`
	Timestamp   time.Time `json:"timestamp,omitempty"`
}

// Bedroom allows the user to update and get information about the bedroom state.
type Bedroom struct {
	db *DB
}

// UpdateTemperature changes bedroom temperature at the specified time, updating the database.
func (b *Bedroom) UpdateTemperature(t time.Time, temp float64) error {
	return b.db.store(t, bedroomField, temp)
}

// FetchState returns the bedroom state updates in the considered period.
func (b *Bedroom) FetchState(start time.Time, finish time.Time) ([]BedroomState, error) {
	iter := b.db.iterRange(start, finish, bedroomField)
	var ret []BedroomState
	var d bedroomTempDocument
	for iter.Next(&d) {
		ret = append(ret, BedroomState{d.Value, d.Timestamp})
	}
	if err := iter.Close(); err != nil {
		if err == mgo.ErrNotFound {
			return ret, nil
		}
		return nil, fmt.Errorf("Error fetching fan status range: %q", err)
	}
	return ret, nil
}

type bedroomTempDocument struct {
	Value     float64   `bson:"value,omitempty"`
	Timestamp time.Time `bson:"timestamp_hour,omitempty"`
}
