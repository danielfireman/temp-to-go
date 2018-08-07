package tsmongo

import (
	"time"
)

const bedroomField = "bedroom"

// BedroomState state stores the bedroom state (e.g. temperature, humidity) at a certain moment.
type BedroomState struct {
	Timestamp   time.Time `json:"timestamp,omitempty"`
	Temperature float64   `json:"temp,omitempty"`
}

// BedroomService allows the user to update and get information about the bedroom state.
type BedroomService struct {
	session *Session
}

// UpdateTemperature changes bedroom temperature at the specified time, updating the database.
func (b *BedroomService) UpdateTemperature(t time.Time, temp float64) error {
	return b.session.Upsert(bedroomField, TSRecord{t, temp})
}

// FetchState returns the bedroom state updates in the considered period.
func (b *BedroomService) FetchState(start time.Time, finish time.Time) ([]BedroomState, error) {
	trs, err := b.session.Query(bedroomField, start, finish)
	if err != nil {
		return nil, err
	}
	ret := make([]BedroomState, len(trs))
	for i := range trs {
		ret[i] = BedroomState{trs[i].Timestamp, trs[i].Value.(float64)}
	}
	return ret, nil
}
