package main

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatalf("Invalid PORT: %s", port)
	}
	key := []byte(os.Getenv("ENCRYPTION_KEY"))
	if len(key) != 32 {
		log.Fatalf("ENCRYPTION_KEY must be 32-bytes long. Current key is \"%s\" which is %d bytes long.", key, len(key))
	}
	router := httprouter.New()
	router.POST("/indoortemp", tempInside(key))
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func tempInside(key []byte) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error reading request body: %q", err), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
		d, err := decrypt(body, key)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error decrypting request body: %q", err), http.StatusForbidden)
			return
		}
		temp, err := strconv.ParseFloat(string(d), 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error request body: %q", err), http.StatusBadRequest)
			return
		}
		fmt.Println(temp)
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
