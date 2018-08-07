package tsmongo

import (
	"fmt"
	"time"

	"github.com/globalsign/mgo"
)

const fanField = "fan"

// FanService allows the user to update and get information about the bedroom fan.
type FanService struct {
	session *Session
}

// ErrInvalidFanStatus represent invalid fan status.
var ErrInvalidFanStatus = fmt.Errorf("Invalid status")

// UpdateStatus changes the fan status at specified time, updating the database.
func (f FanService) UpdateStatus(t time.Time, s FanStatus) error {
	if s < FanOff || s > FanHighSpeed {
		return ErrInvalidFanStatus
	}

	return f.session.Upsert(fanField, TSRecord{time.Now(), s})
}

// LastState returns the last fan status.
func (f FanService) LastState() (FanState, error) {
	ts, err := f.session.Last(fanField)
	switch err {
	case nil:
		return FanState{ts.Timestamp, ts.Value.(FanStatus)}, nil
	case mgo.ErrNotFound:
		return FanState{time.Now(), FanOff}, nil
	default:
		return FanState{time.Now(), FanOff}, err
	}
}

// FetchState returns the fan status updates in the considered period. Important to n
func (f FanService) FetchState(start time.Time, finish time.Time) ([]FanState, error) {
	trs, err := f.session.Query(fanField, start, finish)
	if err != nil {
		return nil, err
	}
	ret := make([]FanState, len(trs))
	for i := range trs {
		ret[i] = FanState{trs[i].Timestamp, trs[i].Value.(FanStatus)}
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
	Timestamp time.Time `json:"timestamp,omitempty"`
	Status    FanStatus `json:"status,omitempty"`
}
