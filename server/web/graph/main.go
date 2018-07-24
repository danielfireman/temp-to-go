package main

//go:generate gopherjs build main.go -o ../public/js/graph.js -m

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/danielfireman/temp-to-go/server/weather"

	charts "github.com/cnguy/gopherjs-frappe-charts"
)

type weatherResponse struct {
	Weather []weather.State `json:"weather,omitempty"`
}

func main() {
	resp, err := http.Get("http://localhost:8080/restricted/weather")
	if err != nil {
		println(err)
		return
	}
	if resp.StatusCode != 200 {
		println("StatusCode:", resp.StatusCode)
		return
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	println(buf.String())

	var wr weatherResponse
	if err := json.Unmarshal(buf.Bytes(), &wr); err != nil {
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
