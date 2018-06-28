package main

import (
	"flag"
	"log"
	"net/http"
	"regexp"
)

var wasmRegex = regexp.MustCompile("\\.wasm$")

func main() {
	d := http.Dir(".")
	fileserver := http.FileServer(d)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ruri := r.RequestURI
		if wasmRegex.MatchString(ruri) {
			w.Header().Set("Content-Type", "application/wasm")
		}
		fileserver.ServeHTTP(w, r)
	})
	addr := flag.String("addr", ":5555", "server address:port")
	flag.Parse()
	log.Printf("listening on %q...", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
