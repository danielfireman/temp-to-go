package main

//go:generate gopherjs build main.go -m -o ../public/js/graph.js

import (
	charts "github.com/cnguy/gopherjs-frappe-charts"
	"github.com/gopherjs/gopherjs/js"

	"honnef.co/go/js/xhr"
)

type tempResponse struct {
	*js.Object
	Hour []string  `js:"hour"`
	Temp []float64 `js:"temp"`
}

const timezoneHeader = "TZ"

func main() {
	// Fetching the timezone.
	tz := js.Global.Get("jstz").Call("determine").Call("name").String()

	// Issuing the request.
	req := xhr.NewRequest("GET", "/restricted/weather")
	req.Timeout = 5000
	req.SetRequestHeader("Content-Type", "application/json")
	req.SetRequestHeader("TZ", tz)
	req.ResponseType = xhr.JSON
	if err := req.Send(nil); err != nil {
		println(err)
		return
	}
	ws := &tempResponse{Object: req.Response}

	reqBS := xhr.NewRequest("GET", "/restricted/indoortemp")
	reqBS.Timeout = 5000
	reqBS.SetRequestHeader("Content-Type", "application/json")
	reqBS.SetRequestHeader("TZ", tz)
	reqBS.ResponseType = xhr.JSON
	if err := reqBS.Send(nil); err != nil {
		println(err)
		return
	}
	bs := &tempResponse{Object: reqBS.Response}

	chartData := charts.NewChartData()
	chartData.Labels = ws.Hour
	chartData.SpecificValues = []*charts.SpecificValue{charts.NewSpecificValue("", "solid", 0)} // Workaround to set the minimum value: https://github.com/frappe/charts/issues/86
	chartData.Datasets = []*charts.Dataset{
		charts.NewDataset(
			"Outdoor Temperatur (Celsius)",
			ws.Temp,
		),
		charts.NewDataset(
			"Indoor Temperature (Celsius)",
			bs.Temp,
		),
	}
	lc := charts.NewLineChart("#chart", chartData)
	lc.RegionFill = 1
	lc.Render()
}
