package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/aio"
	"gobot.io/x/gobot/platforms/firmata"
)

func main() {
	// Validating parameters (environment variables).
	key := []byte(os.Getenv("ENCRYPTION_KEY"))
	if len(key) != 32 {
		log.Fatalf("ENCRYPTION_KEY must be 32-bytes long. Current key is \"%s\" which is %d bytes long.", key, len(key))
	}
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		log.Fatalf("SERVER_URL can not be empty.")
	}
	frequency, err := time.ParseDuration(os.Getenv("FREQUENCY"))
	if err != nil {
		log.Fatalf("Invalid FREQUENCY (\"%s\"):%q", os.Getenv("FREQUENCY"), err)
	}

	// Starting robot.
	firmataAdaptor := firmata.NewTCPAdaptor(os.Args[1])
	tempSensor := aio.NewGroveTemperatureSensorDriver(firmataAdaptor, "3")
	robot := gobot.NewRobot("temperatureRobot",
		[]gobot.Connection{firmataAdaptor},
		[]gobot.Device{tempSensor},
		func() {
			gobot.Every(frequency, func() {
				if err := send(serverURL, tempSensor.Temperature(), key); err != nil {
					log.Println(err)
				}
			})
		},
	)
	robot.Start()
}

var client = &http.Client{
	Timeout: time.Second * 10,
	Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	},
}

func send(u string, temp float64, key []byte) error {
	e, err := encrypt([]byte(strconv.FormatFloat(temp, 'f', -1, 64)), key)
	if err != nil {
		return fmt.Errorf("Error encrypting temperature: %q", err)
	}
	resp, err := client.Post(u, "application/octet-stream", bytes.NewReader([]byte(e)))
	if err != nil {
		return fmt.Errorf("Error trying to send POST request: %q. URL:%s", err, u)
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Invalid status code in POST request: %d. Message: %s", resp.StatusCode, string(b))
	}
	return nil
}

func encrypt(plaintext []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}
