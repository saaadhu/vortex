package proxy

import (
	/*	"github.com/saaadhu/vortex/proxy/cache" */
	"io"
	"log"
	"net"
	"net/http"
	/* "strings" */)

func TunnelTraffic(host string, w http.ResponseWriter) {
	scon, err := net.Dial("tcp", host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	f := w.(http.Flusher)
	f.Flush()

	h, _ := w.(http.Hijacker)
	ccon, _, err := h.Hijack()

	go io.Copy(scon, ccon)
	go io.Copy(ccon, scon)

}

func fetchAndForward(w http.ResponseWriter, r *http.Request) {
	c := http.Client{}
	r.RequestURI = ""
	resp, err := c.Do(r)
	if err != nil {
		log.Fatal(err)
	}

	h, _ := w.(http.Hijacker)
	ccon, bufrw, err := h.Hijack()
	defer ccon.Close()

	resp.Write(bufrw)
	bufrw.Flush()
}

/*
func streamAndCache(w http.ResponseWriter, r *http.Request) {

}
*/

func ProxyTraffic(w http.ResponseWriter, r *http.Request) {

	/*
		if strings.Contains(r.RequestURI, "youtube.com/get_video") {

			v := r.URL.Query()
			id := v.Get("video_id")

			r, err := cache.GetItem(id)
			if err != nil {
				log.Println(err)
				streamAndCache(w, r)
			}

			log.Println(id)

		} else {
	*/
	fetchAndForward(w, r)
}
