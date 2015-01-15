package proxy

import (
	"fmt"
	"github.com/saaadhu/vortex/proxy/cache"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
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

func streamAndCache(id string, w io.Writer, r *http.Request, bRead int64, bTotal int64) {
	c := http.Client{}
	r.RequestURI = ""
	resp, err := c.Do(r)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	log.Println(resp.Header)

	buf := make([]byte, 1*8*1024)
	// We requested a range, server does not support range requests
	// So skip already read bytes
	if bRead > 0 && bTotal > 0 && resp.Header.Get("Content-Range") == "" {
		io.CopyN(ioutil.Discard, resp.Body, bRead)
	}

	if httpw, ok := w.(http.ResponseWriter); ok {
		httpw.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		httpw.Header().Set("Content-Length", resp.Header.Get("Content-Length"))
		httpw.WriteHeader(resp.StatusCode)
	}

	d := make(chan []byte)
	if err := cache.WriteItem(id, resp.Header, d); err != nil {
		log.Fatal(err)
	}
	defer close(d)

	for {
		n, rerr := resp.Body.Read(buf)
		rbuf := buf[:n]

		dbuf := make([]byte, n)
		copy(dbuf, rbuf)
		d <- dbuf

		if _, err := w.Write(rbuf); err != nil {
			log.Println(err)
			break
		}

		if rerr != nil {
			break
		}
	}
}

func serveFromCache(req *http.Request, hr io.Reader, r io.Reader, w http.ResponseWriter) (int64, int64, io.ReadWriter, net.Conn) {
	hi, _ := w.(http.Hijacker)
	ccon, bufrw, _ := hi.Hijack()

	bufrw.WriteString("HTTP/1.1 200 OK\r\n")
	h, err := ioutil.ReadAll(hr)
	if err != nil {
		log.Fatal(err)
	}

	hs := string(h)
	cl := -1
	for _, part := range strings.Split(hs, "\r\n") {
		keyval := strings.Split(part, ":")
		if keyval[0] == "Content-Length" {
			cl, _ = strconv.Atoi(strings.TrimSpace(keyval[1]))
		}
	}

	bufrw.Write(h)
	bufrw.WriteString("\r\n")
	n, _ := io.Copy(bufrw, r)
	bufrw.Flush()

	return int64(cl), n, bufrw, ccon
}

func ProxyTraffic(w http.ResponseWriter, req *http.Request) {

	id := req.RequestURI
	if strings.Contains(req.RequestURI, "c.youtube.com/videoplayback") {
		v := req.URL.Query()
		id = v.Get("id")
	}

	h, r, err := cache.GetItem(id)
	if err != nil {
		streamAndCache(id, w, req, -1, -1)
	} else {
		cl, n, buf, ccon := serveFromCache(req, h, r, w)
		h.Close()
		r.Close()
		if n < cl {
			log.Printf("Requesting range %d-%d", n, cl-1)
			req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", n, cl-1))
			streamAndCache(id, buf, req, n, cl)
		}
		ccon.Close()
	}
}
