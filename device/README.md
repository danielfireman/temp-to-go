# Gotbot + NodeMCU (ESP8266)

* Install Arduino IDE following [the arduino.cc guide](https://www.arduino.cc/en/Guide/Linux)
* Follow the steps described at [How To Connect](https://gobot.io/documentation/platforms/esp8266/#how-to-connect)
     * Install ESP8266 Arduino Addon: OK
     * Download Wifi Enabled Firmata: OK
     * Flash the ESP8266 With Firmata: Caveats
          * You will need to install the ConfigurableFirmata protocol. Instructions under the installation section of [this](https://github.com/firmata/ConfigurableFirmata) page
          * You can have problems with the device or permissions (I had problems writing the firmware to the device). A good set of checks and fixes can be found [here](https://forums.linuxmint.com/viewtopic.php?t=135914)
* Run the [How To Use Program](https://gobot.io/documentation/platforms/esp8266/)
* More resources:
     * [blog.brockman.io::Using Golang With NodeMCU](http://blog.brockman.io/?p=28)
     * [Gobot.io::examples](https://github.com/hybridgroup/gobot/tree/master/examples)
     
# Run the robot

* Install [dep](https://github.com/golang/dep)
* Run the following commands:

```bash
dep init
dep ensure
ENCRYPTION_KEY="the-key-has-to-be-32-bytes-long!" SERVER_URL=http://127.0.0.1:8080/indoortemp FREQUENCY=10s run main.go 192.168.0.67:3030
```
