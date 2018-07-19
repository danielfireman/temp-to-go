package status

import (
	"fmt"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Fan allows the user to update and get information about the bedroom fan.
type Fan struct {
	db *DB
}

// ErrInvalidFanStatus represent invalid fan status.
var ErrInvalidFanStatus = fmt.Errorf("Invalid status")

// UpdateStatus changes the fan status at the current time (now), updating the database.
func (f Fan) UpdateStatus(s FanStatus) error {
	if s < FanOff || s > FanHighSpeed {
		return ErrInvalidFanStatus
	}
	utc := hourUTC(time.Now())
	if s == FanOff {
		err := f.db.collection.Remove(bson.M{timestampIndexField: utc, "type": fanField})
		if err != nil {
			return fmt.Errorf("Error switching the fan off: %q", err)
		}
		return nil
	}
	return f.db.store(time.Now(), fanField, fanState{Value: s})
}

// Status returns the fan status at the current moment.
func (f Fan) Status() (FanStatus, error) {
	utc := hourUTC(time.Now())
	var d fanState
	err := f.db.collection.Find(bson.M{timestampIndexField: utc, "type": fanField}).One(&d)
	switch err {
	case mgo.ErrNotFound:
		return FanOff, nil
	case nil:
		return d.Value, nil
	}
	return FanOff, fmt.Errorf("Error fetching information about the fan: %q", err)
}

// FanStatus represents the speed/power of the bedroom fan.
type FanStatus byte

// Enumerates the three possibilites for the fan speed.
const (
	FanOff       FanStatus = 0
	FanLowSpeed  FanStatus = 1
	FanHighSpeed FanStatus = 2
)

type fanState struct {
	Value FanStatus `bson:"value,omitempty"`
}
