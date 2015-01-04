package cache

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func hashName(name string) string {
	s := sha1.New()
	io.WriteString(s, name)

	return fmt.Sprintf("%x", s.Sum(nil))
}

func GetItem(name string) (io.Reader, io.Reader, error) {
	key := hashName(name)
	log.Printf("Checking if %s in cache", key)
	hfr, err := os.Open(key + ".headers")
	f, err := os.Open(key)
	return hfr, f, err
}

func WriteItem(name string, h http.Header, data chan byte) error {
	key := hashName(name)

	hf, err := os.Create(key + ".headers")
	defer hf.Close()
	h.Write(hf)

	f, err := os.Create(key)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for b := range data {
			if n, err := f.Write([]byte{b}); n != 1 {
				log.Fatal(err)
			}
		}
	}()
	return nil
}
