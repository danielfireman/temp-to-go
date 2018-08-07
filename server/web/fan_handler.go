package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/danielfireman/temp-to-go/server/tsmongo"
	"github.com/labstack/echo"
)

const (
	fanPath            = "/restricted/fan"
	fanStatusFieldName = "fanStatus"
)

type fanHandler struct {
	fanService *tsmongo.FanService
}

// FanHandlerFunc is a handler for APIs relate to the fan.
func (h *fanHandler) handle(c echo.Context) error {
	i, err := strconv.Atoi(c.FormValue(fanStatusFieldName))
	if err != nil {
		c.Logger().Errorf("[/restricted/fan] Invalid fan status: %d\n", i)
		return c.NoContent(http.StatusBadRequest)
	}
	s := tsmongo.FanStatus(byte(i))
	switch h.fanService.UpdateStatus(time.Now(), s) {
	case nil:
		return c.Redirect(http.StatusFound, restrictedPath)
	case tsmongo.ErrInvalidFanStatus:
		c.Logger().Errorf("[/restricted/fan] Invalid fan status: %d %v %s\n", i, s, c.FormValue(fanStatusFieldName))
		return c.NoContent(http.StatusBadRequest)
	default:
		c.Logger().Errorf("[/restricted/fan] %q\n", err)
		return c.NoContent(http.StatusInternalServerError)
	}
}
