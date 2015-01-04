package main

import (
	"github.com/saaadhu/vortex/proxy"
	"log"
	"net/http"
)

func proxyHandler(w http.ResponseWriter, r *http.Request) {

	log.Println(r.URL.String())

	if r.Method == "CONNECT" {
		proxy.TunnelTraffic(r.Host, w)
	} else {
		proxy.ProxyTraffic(w, r)
	}

}

func main() {
	http.ListenAndServe(":3129", http.HandlerFunc(proxyHandler))
}
