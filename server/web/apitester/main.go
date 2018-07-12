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
	"net/http/httputil"
	"os"
	"time"
)

func main() {
	key := []byte(os.Getenv("ENCRYPTION_KEY"))
	if len(key) != 32 {
		log.Fatalf("ENCRYPTION_KEY must be 32-bytes long. Current key is \"%s\" which is %d bytes long.", key, len(key))
	}
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		log.Fatalf("SERVER_URL can not be empty.")
	}

	e, err := encrypt([]byte(os.Args[1]), key)
	if err != nil {
		log.Fatalf("Error encrypting temperature: %q\n", err)
	}
	resp, err := client.Post(serverURL, "application/octet-stream", bytes.NewReader([]byte(e)))
	if err != nil {
		log.Fatalf("Error trying to send POST request: %q. URL:%s\n", err, serverURL)
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		log.Fatalf("Invalid status code in POST request: %d. Message: %s\n", resp.StatusCode, string(b))
	}
	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Fatalf("Error dumping the response: %q", err)
	}
	fmt.Printf("Request sent successfully. Response:\n--\n%s", string(respDump))
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
