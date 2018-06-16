package main

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestSend(t *testing.T) {
	key := []byte("foo-bar-foo-bar-foo-bar-foo-bar-")
	temp := 22.67
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		defer r.Body.Close()
		v, err := decrypt(b, key)
		if err != nil {
			t.Fatal(err)
		}
		got, err := strconv.ParseFloat(string(v), 64)
		if err != nil {
			t.Fatal(err)
		}
		if temp != got {
			t.Fatalf("want:%f got:%f", temp, got)
		}
	}))
	defer ts.Close()
	if err := send(ts.URL, temp, key); err != nil {
		t.Fatalf("error sending mensage: %q", err)
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
