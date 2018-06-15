package main

import (
	"fmt"
	"os"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/aio"
	"gobot.io/x/gobot/platforms/firmata"
)

func main() {
	firmataAdaptor := firmata.NewTCPAdaptor(os.Args[1])
	temp := aio.NewGroveTemperatureSensorDriver(firmataAdaptor, "3")
	work := func() {
		gobot.Every(1*time.Second, func() {
			fmt.Printf("Temp: %f\n", temp.Temperature())
		})
	}
	robot := gobot.NewRobot("bot",
		[]gobot.Connection{firmataAdaptor},
		[]gobot.Device{temp},
		work,
	)
	robot.Start()
}
