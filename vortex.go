package main

import (
	"flag"
	"github.com/saaadhu/vortex/proxy"
	"github.com/saaadhu/vortex/proxy/cache"
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
	cd := flag.String("cachedir", "", "Cache directory")
	flag.Parse()
	log.Printf("Vortex starting with cache at %s", *cd)

	cache.Init(*cd)
	http.ListenAndServe(":3129", http.HandlerFunc(proxyHandler))
}
