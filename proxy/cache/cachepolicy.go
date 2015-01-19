package cache

import (
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func isStale(key string) bool {

	hfp, _ := getFilePaths(key)
	h, err := ioutil.ReadFile(hfp)
	if err != nil {
		return false
	}

	hs := string(h)
	maxAge := -1
	serverDate := ""
	contentType := ""
	re := regexp.MustCompile("max-age=([0-9]+)")
	for _, part := range strings.Split(hs, "\r\n") {
		keyval := strings.Split(part, ":")
		if keyval[0] == "Cache-Control" {
			if res := re.FindStringSubmatch(keyval[1]); res != nil {
				maxAge, _ = strconv.Atoi(res[1])
			}
		} else if keyval[0] == "Date" {
			serverDate = strings.TrimSpace(strings.Join(keyval[1:], ":"))
		} else if keyval[0] == "Content-Type" {
			contentType = strings.TrimSpace(keyval[1])
		}
	}
	log.Println(maxAge, serverDate, contentType)

	if maxAge == -1 || serverDate == "" {
		return false
	}

	if m, _ := regexp.MatchString("video|audio|image", contentType); m {
		return false
	}

	t, err := time.Parse(time.RFC1123, serverDate)
	if err != nil {
		log.Fatal(err)
	}
	return t.Add(time.Duration(maxAge) * time.Second).Before(time.Now())
}
