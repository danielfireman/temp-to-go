package weather

// State stores the complete information about the weather at a certain time.
type State struct {
	Timestamp   int64       // Timestamp in unix UTC
	Description Description // Icon and other text describing the weather state
	Wind        Wind        // Wind description
	Temp        float32     // Temperature, Celsius
	Humidity    float32     // Humidity, %
	Rain        float32     // Rain volume for the last hours
	Cloudiness  float32     // Cloudiness, %
}

// Wind stores information about the wind.
type Wind struct {
	Speed     float32 `bson:"speed,omitempty"` // Wind speed, meter/sec
	Direction float32 `bson:"deg,omitempty"`   // Wind direction, degrees (meteorological)
}

// Description stores overall information to describe the weather. That include text, images and so on.
type Description struct {
	Text string `bson:"text,omitempty"` // Weather condition within the group
	Icon string `bson:"icon,omitempty"` // Weather icon id
}
