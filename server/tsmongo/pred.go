package tsmongo

import "time"

const predictionField = "pred"

// PredictionService allows users to store or fetch predictions.
type PredictionService struct {
	session *Session
}

// Update stores the passed-in predictions. The call overrides any
// previously stored predictions.
func (p *PredictionService) Update(ps ...Prediction) error {
	if len(ps) == 0 {
		return nil
	}
	trs := make([]TSRecord, len(ps))
	for i := range ps {
		trs[i] = TSRecord{ps[i].Timestamp, ps[i]}
	}
	return p.session.Upsert(predictionField, trs...)
}

// Prediction represents a prediction of a bedroom temperature at a certain time in
// any of the following states: fan off, low or high.
type Prediction struct {
	Timestamp   time.Time `bson:"-"`
	TempFanOff  float64   `bson:"fan_off,omitempty"`
	TempFanLow  float64   `bson:"fan_low,omitempty"`
	TempFanHigh float64   `bson:"fan_high,omitempty"`
}
