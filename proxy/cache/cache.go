package cache

import (
	"crypto/sha1"
	"errors"
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

func getFilePaths(key string) (string, string) {
	p := fmt.Sprintf("%s%c%s", cacheDir, os.PathSeparator, key)
	return p + ".headers", p
}

func removeItem(key string) {
	p1, p2 := getFilePaths(key)
	os.Remove(p1)
	os.Remove(p2)
}

func GetItem(name string) (io.ReadWriteCloser, io.ReadWriteCloser, error) {
	key := hashName(name)

	if isStale(key) {
		removeItem(key)
		return nil, nil, errors.New("Cache item expired")
	}

	hfp, fp := getFilePaths(key)
	hfr, err := os.OpenFile(hfp, os.O_RDWR|os.O_APPEND, os.ModePerm)
	f, err := os.OpenFile(fp, os.O_RDWR|os.O_APPEND, os.ModePerm)
	return hfr, f, err
}

func WriteItem(name string, h http.Header, data chan []byte) error {
	key := hashName(name)

	hfp, fp := getFilePaths(key)
	hf, f, err := GetItem(name)
	if err != nil {
		hf, err = os.Create(hfp)

		h.Write(hf)

		f, err = os.Create(fp)
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
