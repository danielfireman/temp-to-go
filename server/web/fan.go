package main

import (
	"net/http"
	"strconv"

	"github.com/danielfireman/temp-to-go/server/status"
	"github.com/labstack/echo"
)

// FanHandlerFunc is a handler for APIs relate to the fan.
func fanHandlerFunc(fan *status.Fan) echo.HandlerFunc {
	return func(c echo.Context) error {
		i, err := strconv.Atoi(c.FormValue(fanStatusFieldName))
		if err != nil {
			c.Logger().Errorf("[/restricted/fan] Invalid fan status: %d\n", i)
			return c.NoContent(http.StatusBadRequest)
		}
		s := status.FanStatus(byte(i))
		switch fan.UpdateStatus(s) {
		case nil:
			return c.Redirect(http.StatusFound, restrictedPath)
		case status.ErrInvalidFanStatus:
			c.Logger().Errorf("[/restricted/fan] Invalid fan status: %d %v %s\n", i, s, c.FormValue(fanStatusFieldName))
			return c.NoContent(http.StatusBadRequest)
		default:
			c.Logger().Errorf("[/restricted/fan] %q\n", err)
			return c.NoContent(http.StatusInternalServerError)
		}
	}
}