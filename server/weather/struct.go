package weather

import "time"

// State stores the complete information about the weather at a certain time.
type State struct {
	Timestamp   time.Time   // Timestamp in unix UTC
	Description Description // Icon and other text describing the weather state
	Wind        Wind        // Wind description
	Temp        float64     // Temperature, Celsius
	Humidity    float64     // Humidity, %
	Rain        float64     // Rain volume for the last hours
	Cloudiness  float64     // Cloudiness, %
}

// Wind stores information about the wind.
type Wind struct {
	Speed     float64 `bson:"speed,omitempty"` // Wind speed, meter/sec
	Direction float64 `bson:"deg,omitempty"`   // Wind direction, degrees (meteorological)
}

// Description stores overall information to describe the weather. That include text, images and so on.
type Description struct {
	Text string `bson:"text,omitempty"` // Weather condition within the group
	Icon string `bson:"icon,omitempty"` // Weather icon id
}
