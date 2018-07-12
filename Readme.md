# Temperature to go

Would you always wanted to know when to turn on the fan or the air conditioner during the night? Good news is that your problem will be solved soon!

Besides solve this very important problem, our small projects has some other goals:

* Fully implemented in Go (yes, from the device to frontend)
* Use only free, open source libraries
* Cheap devices
* Run on free infrastructure

# Status

* Device: Done :grin:
* HTTP API (bedroom temperature): Done :grin:
* Fetch Current Weather Worker: Started :smirk:
* Fetch Weather Forecast and Prediction Worker: Not started :unamused:
* Fan control HTTP FE: Not Started :unamused:
* Public HTTP FE: Not started :unamused:

# Design

> TODO

# Stack

* Device/Sensor
     * Hardware
          * [NodeMCU / ESP8266](http://nodemcu.com/index_en.html)
          * [Grove Temperature Sensor](http://wiki.seeedstudio.com/Grove-Temperature_Sensor_V1.2/)
     * Software
          * Development using [gobot.io](https://gobot.io/documentation/platforms/esp8266/) through [firmata protocol](https://github.com/firmata/arduino)

* Server/Worker
     * Platform
          * [Heroku](https://www.heroku.com/): Cloud provider
          * [Namecheap](https://www.namecheap.com/): Domain names (mybedroom.live and meuquarto.live)
          * [Cloudflare](https://www.cloudflare.com/): Better DNS, DoS protection, edge caching and more
          * [MongoDB](https://www.mongodb.com/): [mongoLab Addon](https://elements.heroku.com/addons/mongolab)
     * Software
          * [mgo](https://labix.org/mgo)
          * [Gonum](https://github.com/gonum/gonum), in particular the [Linear Regression](https://godoc.org/gonum.org/v1/gonum/stat#LinearRegression)
          

* Client
     * [GopherJS](https://github.com/gopherjs/gopherjs)
     * [Frappé-charts](https://github.com/cnguy/gopherjs-frappe-charts)
    
