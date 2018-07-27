package main

//go:generate gopherjs build main.go -m -o ../public/js/graph.js

import (
	"github.com/cathalgarvey/fmtless/encoding/json"
	charts "github.com/cnguy/gopherjs-frappe-charts"
	"github.com/gopherjs/gopherjs/js"

	"honnef.co/go/js/xhr"
)

type weatherResponse struct {
	Hour []string  `json:"hour,omitempty"`
	Temp []float64 `json:"temp,omitempty"`
}

const timezoneHeader = "TZ"

func main() {
	// Fetching the timezone.
	tz := js.Global.Get("jstz").Call("determine").Call("name").String()

	// Issuing the request.
	req := xhr.NewRequest("GET", "/restricted/weather")
	req.Timeout = 5000
	req.ResponseType = xhr.ArrayBuffer
	req.SetRequestHeader("TZ", tz)
	if err := req.Send(tz); err != nil {
		println(err)
		return
	}
	if req.Status != 200 {
		println("Not 200")
		return
	}
	b := js.Global.Get("Uint8Array").New(req.Response).Interface().([]byte)
	println(string(b))
	var wr weatherResponse
	if err := json.Unmarshal(b, &wr); err != nil {
		println(err)
		return
	}
	chartData := charts.NewChartData()
	chartData.Labels = wr.Hour
	chartData.SpecificValues = []*charts.SpecificValue{charts.NewSpecificValue("", "solid", 0)}
	// Workaround to set the minimum value: https://github.com/frappe/charts/issues/86
	chartData.Datasets = []*charts.Dataset{
		charts.NewDataset(
			"Temperature (Celsius)",
			wr.Temp,
		)}
	lc := charts.NewLineChart("#chart", chartData)
	lc.RegionFill = 1
	lc.Render()
}
