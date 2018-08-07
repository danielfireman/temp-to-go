package main

import (
	"net/http"

	"github.com/danielfireman/temp-to-go/server/tsmongo"
	"github.com/labstack/echo"
)

type restrictedMainHandler struct {
	fan *tsmongo.FanService
}

func (h *restrictedMainHandler) handle(c echo.Context) error {
	s, err := h.fan.LastState()
	if err != nil {
		c.Logger().Errorf("[main] %q\n", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// Converting the FanSpeed to text.
	currSpeed := "Off"
	switch s.Status {
	case tsmongo.FanLowSpeed:
		currSpeed = "Low"
	case tsmongo.FanHighSpeed:
		currSpeed = "High"
	}

	// Struct containing options to draw the radio button options.
	type fanOpt struct {
		Label string
		Value tsmongo.FanStatus
		Name  string
	}
	return c.Render(http.StatusOK, "main", struct {
		Speed  string
		Opts   []fanOpt
		Action string
	}{
		Speed: currSpeed,
		Opts: []fanOpt{
			{"Off", tsmongo.FanOff, fanStatusFieldName},
			{"Low", tsmongo.FanLowSpeed, fanStatusFieldName},
			{"High", tsmongo.FanHighSpeed, fanStatusFieldName},
		},
		Action: fanPath,
	})
}
