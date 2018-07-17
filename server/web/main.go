package main

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/danielfireman/temp-to-go/server/status"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"golang.org/x/crypto/acme/autocert"
)

// Defines the environment.
var env = os.Getenv("ENV")

func main() {
	key := []byte(os.Getenv("ENCRYPTION_KEY"))
	if len(key) != 32 {
		log.Fatalf("ENCRYPTION_KEY must be 32-bytes long. Current key is \"%s\" which is %d bytes long.", key, len(key))
	}
	mgoURI := os.Getenv("MONGODB_URI")
	if mgoURI == "" {
		log.Fatalf("Invalid MONGODB_URI: %s", mgoURI)
	}
	sdb, err := status.DialDB(mgoURI)
	if err != nil {
		log.Fatalf("Error connecting to StatusDB: %s", mgoURI)
	}
	defer sdb.Close()

	e := echo.New()

	env := os.Getenv("ENV")
	if env == "PROD" {
		// Configuring TLS.
		log.Println(strings.Split(os.Getenv("TLS_HOST_WHITELIST"), ","))
		e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(strings.Split(os.Getenv("TLS_HOST_WHITELIST"), ",")...)
		tlsCachePath, err := ioutil.TempDir("", ".tlscache")
		if err != nil {
			log.Fatalf("Could not create TLS temporary cache directory: %q", err)
		}
		e.AutoTLSManager.Cache = autocert.DirCache(tlsCachePath)
	}

	// Middlewares.
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	// Paths.
	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, `
			<h1>Welcome to MyBedroom!</h1>
			<h3>TLS certificates automatically installed from Let's Encrypt :)</h3>
		`)
	})
	e.POST("/indoortemp", indoorTemp(key, sdb))

	if isProdEnv() {
		e.Logger.Fatal(e.StartAutoTLS(":443"))
	} else {
		port := os.Getenv("PORT")
		if port == "" {
			log.Fatalf("Invalid PORT: %s", port)
		}
		s := &http.Server{
			Addr:         ":" + port,
			ReadTimeout:  5 * time.Minute,
			WriteTimeout: 5 * time.Minute,
		}
		e.Logger.Fatal(e.StartServer(s))
	}
}

func isProdEnv() bool {
	return env == "PROD"
}

func indoorTemp(key []byte, db *status.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var body []byte
		if err := c.Bind(&body); err != nil {
			c.Logger().Errorf("Error reading request body: %q", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		d, err := decrypt(body, key)
		if err != nil {
			c.Logger().Errorf("Error decrypting request body: %q", err)
			return c.NoContent(http.StatusForbidden)
		}
		temp, err := strconv.ParseFloat(string(d), 32)
		if err != nil {
			c.Logger().Errorf("Error request body: %q", err)
			return c.NoContent(http.StatusBadRequest)
		}
		if err := db.StoreBedroomTemperature(time.Now(), float32(temp)); err != nil {
			c.Logger().Errorf("StoreBedroomTemperature: %q\n", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		return nil
	}
}

func decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
