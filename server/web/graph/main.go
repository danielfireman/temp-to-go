package main

//go:generate gopherjs build main.go -m -o ../public/js/graph.js

import (
	"strconv"
	"time"

	"github.com/cathalgarvey/fmtless/encoding/json"
	charts "github.com/cnguy/gopherjs-frappe-charts"
	"github.com/gopherjs/gopherjs/js"

	"honnef.co/go/js/xhr"
)

type weatherResponse struct {
	Weather []struct {
		Timestamp time.Time // Timestamp in unix UTC
		Temp      float64   // Temperature, Celsius
	} `json:"weather,omitempty"`
}

func main() {
	req := xhr.NewRequest("GET", "http://localhost:8080/restricted/weather")
	req.Timeout = 5000
	req.ResponseType = xhr.ArrayBuffer
	if err := req.Send(nil); err != nil {
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
	var temps []float64
	var labels []string
	for _, s := range wr.Weather {
		d := s.Timestamp.Hour()
		labels = append(labels, strconv.Itoa(d))
		temps = append(temps, s.Temp)
	}
	chartData := charts.NewChartData()
	chartData.Labels = labels
	chartData.Datasets = []*charts.Dataset{
		charts.NewDataset(
			"Weather",
			temps,
		)}

	_ = charts.NewLineChart("#chart", chartData).Render()
}
