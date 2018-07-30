package status

import (
	"fmt"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const fanField = "fan"

// Fan allows the user to update and get information about the bedroom fan.
type Fan struct {
	db *DB
}

// ErrInvalidFanStatus represent invalid fan status.
var ErrInvalidFanStatus = fmt.Errorf("Invalid status")

// UpdateStatus changes the fan status at specified time, updating the database.
func (f Fan) UpdateStatus(t time.Time, s FanStatus) error {
	if s < FanOff || s > FanHighSpeed {
		return ErrInvalidFanStatus
	}
	utc := hourUTC(t)
	if s == FanOff {
		err := f.db.collection.Remove(bson.M{timestampIndexField: utc, "type": fanField})
		if err != nil {
			return fmt.Errorf("Error switching the fan off: %q", err)
		}
		return nil
	}
	return f.db.store(time.Now(), fanField, s)
}

// LastState returns the last fan status.
func (f Fan) LastState() (FanState, error) {
	var d fanDocument
	err := f.db.collection.Find(bson.M{"type": fanField}).Sort("-" + timestampIndexField).One(&d)
	if err != nil {
		return FanState{FanOff, time.Now().In(time.UTC)}, fmt.Errorf("There isError fetching information about the fan: %q", err)
	}
	return FanState{d.Value, d.Timestamp}, nil
}

// FetchState returns the fan status updates in the considered period. Important to n
func (f Fan) FetchState(start time.Time, finish time.Time) ([]FanState, error) {
	iter := f.db.iterRange(start, finish, fanField)
	var ret []FanState
	var d fanDocument
	for iter.Next(&d) {
		ret = append(ret, FanState{d.Value, d.Timestamp})
	}
	if err := iter.Close(); err != nil {
		if err == mgo.ErrNotFound {
			return ret, nil
		}
		return nil, fmt.Errorf("Error fetching fan status range: %q", err)
	}
	return ret, nil
}

// FanStatus represents the speed/power of the bedroom fan.
type FanStatus byte

// Enumerates the three possibilites for the fan speed.
const (
	FanOff       FanStatus = 0
	FanLowSpeed  FanStatus = 1
	FanHighSpeed FanStatus = 2
)

// FanState stores the state of the fan at a certain moment.
type FanState struct {
	Status    FanStatus `json:"status,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// Represents the actual document stored in the database. For now, we only care about the value field.
type fanDocument struct {
	Value     FanStatus `bson:"value,omitempty"`
	Timestamp time.Time `bson:"timestamp_hour,omitempty"`
}
