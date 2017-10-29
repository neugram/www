package main

//go:generate go run genstatic.go -o static.go

import (
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"encoding/base64"
	"flag"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

var staticHTML = make(map[string]string)

var start = time.Now()
var addr = flag.String("addr", ":443", "address to listen on")

func makeEtag(b []byte) string {
	sum := sha1.Sum(b)
	return `"` + base64.StdEncoding.EncodeToString(sum[:]) + `"`
}

func staticHandler(path, gzb64src string) func(w http.ResponseWriter, req *http.Request) {
	gzsrc, err := base64.StdEncoding.DecodeString(gzb64src)
	if err != nil {
		log.Fatalf("error decoding %s: %v", path, err)
	}

	r, err := gzip.NewReader(bytes.NewReader(gzsrc))
	if err != nil {
		log.Fatalf("error gunzipping %s: %v", path, err)
	}
	src, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatalf("error gunzipping %s: %v", path, err)
	}

	ctype := mime.TypeByExtension(filepath.Ext(path))
	if ctype == "" {
		ctype = http.DetectContentType(src)
	}
	gzetag := makeEtag(gzsrc)
	etag := makeEtag(src)

	return func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != path {
			http.NotFound(w, req)
			return
		}
		w.Header().Set("Content-Type", ctype)

		if strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Etag", gzetag)
			http.ServeContent(w, req, path, start, bytes.NewReader(gzsrc))
			return
		}

		w.Header().Set("Etag", etag)
		http.ServeContent(w, req, path, start, bytes.NewReader(src))
	}
}

func main() {
	flag.Parse()

	for name, src := range staticHTML {
		http.HandleFunc(name, staticHandler(name, src))
	}

	log.Printf("serving neugram.io on %s", *addr)
	if *addr == ":443" {
		log.Fatal(http.ListenAndServeTLS(*addr, "www.pem", "www.key", nil))
	} else {
		log.Fatal(http.ListenAndServe(*addr, nil))
	}

}
