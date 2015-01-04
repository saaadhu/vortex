package cache

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"os"
)

func hashName(name string) string {
	s := sha1.New()
	io.WriteString(s, name)

	return fmt.Sprintf("%x", s.Sum(nil))
}

func GetItem(name string) (io.Reader, error) {
	key := hashName(name)
	log.Printf("Checking if %s in cache", key)
	return os.Open(key)
}

func WriteItem(name string, data chan byte) error {
	key := hashName(name)

	f, err := os.Create(key)
	if err != nil {
		return err
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
