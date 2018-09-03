package main

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/danielfireman/temp-to-go/server/tsmongo"
	"github.com/labstack/echo"
)

type bedroomAPIHandler struct {
	key            []byte
	bedroomService *tsmongo.BedroomService
}

func (h *bedroomAPIHandler) handle(c echo.Context) error {
	var body []byte
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		c.Logger().Errorf("Error reading request body: %q", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	defer c.Request().Body.Close()
	d, err := decrypt(body, h.key)
	if err != nil {
		c.Logger().Errorf("Error decrypting request body: %q", err)
		return c.NoContent(http.StatusForbidden)
	}
	temp, err := strconv.ParseFloat(string(d), 64)
	if err != nil {
		c.Logger().Errorf("Error request body: %q", err)
		return c.NoContent(http.StatusBadRequest)
	}
	if err := h.bedroomService.UpdateTemperature(time.Now(), temp); err != nil {
		c.Logger().Errorf("StoreBedroomTemperature: %q\n", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return nil
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
