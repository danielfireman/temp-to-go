package weather

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"time"
)

const (
	owmForecastCall    = "forecast"
	owmCurrWeatherCall = "weather"
	owmFmtURL          = "http://api.openweathermap.org/data/2.5/%s?units=metric&q=Maceio,BR&appid=%s"
)

// Client is used to communcate with a service and fetch current and future information about the weather.
type Client struct {
	c   *http.Client
	key string
}

// NewOWMClient a new Client, which talks to the OpenWeatherMaps (https://openweathermap.org).
func NewOWMClient(key string) *Client {
	return &Client{
		c: &http.Client{
			Timeout: time.Second * 10,
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).Dial,
				TLSHandshakeTimeout: 5 * time.Second,
			},
		},
		key: key,
	}
}

// Current fetches and returns the current weather state from open weather maps.
func (c Client) Current() (State, error) {
	resp, err := c.c.Get(fmt.Sprintf(owmFmtURL, owmCurrWeatherCall, c.key))
	if err != nil {
		return State{}, fmt.Errorf("Error fetching current weather from OWM: %q", err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return State{}, fmt.Errorf("Error reading OMW response body: %q", err)
	}
	var owmResp owmCurrWeatherResponse
	if err := json.Unmarshal(b, &owmResp); err != nil {
		return State{}, fmt.Errorf("Error unmarshalling OMW response: %q", err)
	}
	return toState(owmResp), nil
}

// Forecast fetches and returns the weather forecast. As the forecast grain (e.g., each 3 hours for the next day) can
// vary and callers need to inspect each of the State's timestamp to identify it.
func (c Client) Forecast() ([]State, error) {
	resp, err := c.c.Get(fmt.Sprintf(owmFmtURL, owmCurrWeatherCall, c.key))
	if err != nil {
		return nil, fmt.Errorf("Error fetching current weather from OWM: %q", err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading OMW response body: %q", err)
	}
	var owmForecastResp owmForecastWeatherResponse
	if err := json.Unmarshal(b, &owmForecastResp); err != nil {
		return nil, fmt.Errorf("Error unmarshalling OMW response: %q", err)
	}
	var states []State
	for _, owmResp := range owmForecastResp.List {
		states = append(states, toState(owmResp))
	}
	return states, nil
}

func toState(owmResp owmCurrWeatherResponse) State {
	return State{
		Description: Description{
			Text: owmResp.Weather[0].Description,
			Icon: owmResp.Weather[0].Icon,
		},
		Wind: Wind{
			Speed:     round(owmResp.Wind.Speed),
			Direction: round(owmResp.Wind.Deg),
		},
		Temp:       round(owmResp.Main.Temp),
		Humidity:   round(owmResp.Main.Humidity),
		Rain:       round(owmResp.Rain.ThreeHours),
		Cloudiness: round(owmResp.Clouds.All),
		Timestamp:  owmResp.DT,
	}
}

// Structs used in the communication with OWM.
type owmForecastWeatherResponse struct {
	List []owmCurrWeatherResponse `json:"list,omitempty"`
}

type owmCurrWeatherResponse struct {
	Weather []weather `json:"weather,omitempty"`
	Main    mainn     `json:"main,omitempty"`
	Clouds  clouds    `json:"clouds,omitempty"`
	Wind    wind      `json:"wind,omitempty"`
	Rain    rain      `json:"rain,omitempty"`
	DT      int64     `json:"dt,omitempty"` // Time of data calculation, unix, UTC
}

type clouds struct {
	All float64 `json:"all,omitempty"` //  Cloudiness, %
}

type mainn struct {
	Temp     float64 `json:"temp,omitempty"`     // Temperature, Celsius
	Humidity float64 `json:"humidity,omitempty"` // Humidity, %
}

type wind struct {
	Speed float64 `json:"speed,omitempty"` // Wind speed, meter/sec
	Deg   float64 `json:"deg,omitempty"`   // Wind direction, degrees (meteorological)
}

type rain struct {
	ThreeHours float64 `json:"3h,omitempty"` // Rain volume for the last 3 hours
}

type weather struct {
	Description string `json:"description,omitempty"` // Weather condition within the group
	Icon        string `json:"icon,omitempty"`        // Weather icon id
}

// round receives a float64, rounds it to 4 most signigicant digits (max) and returns it as
// float32. Mostly used to decrease massive (unnecessary) precision of float64 and thus
// to decrease storage requirements.
func round(f float64) float32 {
	return float32(math.Round(f/0.0001) * 0.0001)
}
