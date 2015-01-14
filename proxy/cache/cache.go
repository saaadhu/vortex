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

var cacheDir string

func Init(cd string) {
	cacheDir = cd
}

func GetItem(name string) (io.ReadWriteCloser, io.ReadWriteCloser, error) {
	key := hashName(name)
	log.Printf("Checking if %s in cache", key)
	hfr, err := os.OpenFile(cacheDir+"/"+key+".headers", os.O_RDWR|os.O_APPEND, os.ModePerm)
	f, err := os.OpenFile(cacheDir+"/"+key, os.O_RDWR|os.O_APPEND, os.ModePerm)
	return hfr, f, err
}

func WriteItem(name string, h http.Header, data chan []byte) error {
	key := hashName(name)

	hf, f, err := GetItem(name)
	if err != nil {
		hf, err = os.Create(cacheDir + "/" + key + ".headers")

		h.Write(hf)

		f, err = os.Create(cacheDir + "/" + key)
		if err != nil {
			log.Fatal(err)
		}
	}
	hf.Close()

	go func() {
		for b := range data {
			if _, err := f.Write(b); err != nil {
				log.Fatal(err)
			}
		}
		f.Close()
	}()
	return nil
}
