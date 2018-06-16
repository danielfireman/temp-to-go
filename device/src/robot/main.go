package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/aio"
	"gobot.io/x/gobot/platforms/firmata"
)

func main() {
	key := []byte(os.Getenv("ENCRYPTION_KEY"))
	if len(key) != 32 {
		log.Fatalf("ENCRYPTION_KEY must be 32-bytes long. Current key is \"%s\" which is %d bytes long.", key, len(key))
	}
	firmataAdaptor := firmata.NewTCPAdaptor(os.Args[1])
	temp := aio.NewGroveTemperatureSensorDriver(firmataAdaptor, "3")
	work := func() {
		gobot.Every(1*time.Second, func() {
			temp := temp.Temperature()
			eTemp, err := encrypt([]byte(strconv.FormatFloat(temp, 'f', -1, 64)), key)
			if err != nil {
				log.Fatalf("Error trying to encrypt temperature (%f): %q", temp, err)
			}
			fmt.Printf("Temp: %f, Encrypted: %s\n", temp, eTemp)
		})
	}
	robot := gobot.NewRobot("bot",
		[]gobot.Connection{firmataAdaptor},
		[]gobot.Device{temp},
		work,
	)
	robot.Start()
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
