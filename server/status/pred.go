package status

import "time"

const predictionField = "pred"

// Predictions allows users to store or fetch predictions.
type Predictions struct {
	db *DB
}

// Update stores the passed-in predictions. The call overrides any
// previously stored predictions.
func (p *Predictions) Update(ps ...Prediction) error {
	bulk := p.db.collection.Bulk()
	for _, s := range ps {
		p.db.bulkStore(bulk, s.Timestamp, predictionField, ps)
	}
	_, err := bulk.Run()
	return err
}

// Prediction represents a prediction of a bedroom temperature at a certain time in
// any of the following states: fan off, low or high.
type Prediction struct {
	Timestamp   time.Time `bson:"-"`
	TempFanOff  float64   `bson:"fan_off,omitempty"`
	TempFanLow  float64   `bson:"fan_low,omitempty"`
	TempFanHigh float64   `bson:"fan_high,omitempty"`
}

type predictionDocument struct {
	Timestamp time.Time  `bson:"timestamp_hour,omitempty"`
	Value     Prediction `bson:"value,omitempty"`
}
