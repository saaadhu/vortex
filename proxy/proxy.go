package proxy

import (
	"github.com/saaadhu/vortex/proxy/cache"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

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

func streamAndCache(id string, w http.ResponseWriter, r *http.Request) {
	c := http.Client{}
	r.RequestURI = ""
	resp, err := c.Do(r)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))
	w.WriteHeader(resp.StatusCode)

	d := make(chan byte, 5*1024*1024)
	if err := cache.WriteItem(id, resp.Header, d); err != nil {
		log.Fatal(err)
	}
	defer close(d)

	buf := make([]byte, 5*1024*1024)

	for {
		n, err := resp.Body.Read(buf)
		log.Println(n)
		w.Write(buf[:n])

		for i := 0; i < n; i = i + 1 {
			d <- buf[i]
		}

		if err != nil {
			break
		}
	}
}

func serveFromCache(hr io.Reader, r io.Reader, w http.ResponseWriter) {
	h, _ := w.(http.Hijacker)
	ccon, bufrw, _ := h.Hijack()
	defer ccon.Close()

	io.Copy(w, hr)
	io.Copy(w, r)
	bufrw.Flush()
}

func ProxyTraffic(w http.ResponseWriter, req *http.Request) {

	if strings.Contains(req.RequestURI, "c.youtube.com/videoplayback") {

		v := req.URL.Query()
		id := v.Get("id")

		h, r, err := cache.GetItem(id)
		if err != nil {
			streamAndCache(id, w, req)
		} else {
			serveFromCache(h, r, w)
		}
	} else {
		fetchAndForward(w, req)
	}
}
