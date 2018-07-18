package bedroomapi

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/danielfireman/temp-to-go/server/status"
	"github.com/labstack/echo"
)

// TempHandlerFunc returns a echo.HandlerFunc which can handle call to the BedroomTemp API.
func TempHandlerFunc(key []byte, db *status.DB) echo.HandlerFunc {
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
