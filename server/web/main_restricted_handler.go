package main

import (
	"net/http"

	"github.com/danielfireman/temp-to-go/server/status"
	"github.com/labstack/echo"
)

type restrictedMainHandler struct {
	fan *status.Fan
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
	case status.FanLowSpeed:
		currSpeed = "Low"
	case status.FanHighSpeed:
		currSpeed = "High"
	}

	// Struct containing options to draw the radio button options.
	type fanOpt struct {
		Label string
		Value status.FanStatus
		Name  string
	}
	return c.Render(http.StatusOK, "main", struct {
		Speed  string
		Opts   []fanOpt
		Action string
	}{
		Speed: currSpeed,
		Opts: []fanOpt{
			{"Off", status.FanOff, fanStatusFieldName},
			{"Low", status.FanLowSpeed, fanStatusFieldName},
			{"High", status.FanHighSpeed, fanStatusFieldName},
		},
		Action: fanPath,
	})
}
