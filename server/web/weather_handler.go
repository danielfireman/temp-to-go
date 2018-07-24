package main

import (
	"net/http"
	"time"

	"github.com/danielfireman/temp-to-go/server/status"
	"github.com/danielfireman/temp-to-go/server/weather"
	"github.com/labstack/echo"
)

func weatherHandlerFunc(db *status.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var resp weatherResponse
		var err error
		resp.Weather, err = db.FetchWeather(time.Now().Add(-24*time.Hour), time.Now())
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.JSON(http.StatusOK, resp)
	}
}

type weatherResponse struct {
	Weather []weather.State `json:"weather,omitempty"`
}
