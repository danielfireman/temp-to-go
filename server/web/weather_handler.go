package main

import (
	"net/http"
	"time"

	"github.com/danielfireman/temp-to-go/server/tsmongo"
	"github.com/labstack/echo"
)

type weatherHandler struct {
	weatherService *tsmongo.WeatherService
}

const timezoneHeader = "TZ"

func (h *weatherHandler) handle(c echo.Context) error {
	var err error
	ws, err := h.weatherService.Fetch(time.Now().Add(-24*time.Hour), time.Now())
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	loc, err := time.LoadLocation(c.Request().Header.Get(timezoneHeader))
	if err != nil {
		c.Logger().Error(err)
		loc = time.UTC
	}
	var resp weatherResponse
	for _, s := range ws {
		t := s.Timestamp.In(loc)
		resp.Hour = append(resp.Hour, t.Format("3pm"))
		resp.Temp = append(resp.Temp, s.Temp)
	}
	return c.JSON(http.StatusOK, resp)
}

type weatherResponse struct {
	Hour []string  `json:"hour,omitempty"`
	Temp []float64 `json:"temp,omitempty"`
}
