// +build ignore

package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var outfile = flag.String("o", "", "result will be written file")

func writeFile(buf *bytes.Buffer, name string, path string) {
	log.Printf("writing %s", name)
	src, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	gzbuf := new(bytes.Buffer)
	gzw, err := gzip.NewWriterLevel(gzbuf, gzip.BestCompression)
	if err != nil {
		log.Fatal(err)
	}
	gzw.Write(src)
	if err := gzw.Close(); err != nil {
		log.Fatal(err)
	}
	data := base64.StdEncoding.EncodeToString(gzbuf.Bytes())

	fmt.Fprintf(buf, "func init() {\n\tstaticHTML[%q] = `` + \n", name)
	chunk := ""
	for len(data) > 0 {
		l := len(data)
		if l > 72 {
			l = 72
		}
		chunk, data = data[:l], data[l:]
		fmt.Fprintf(buf, "\t`%s` + \n", chunk)
	}
	fmt.Fprintf(buf, "\t``\n}")
}

func main() {
	flag.Parse()

	buf := new(bytes.Buffer)
	fmt.Fprint(buf, "package main\n\n")

	writeFile(buf, "/", "main.html")

	blogFiles, err := filepath.Glob("blog/*")
	if err != nil {
		log.Fatal(err)
	}
	for _, blogFile := range blogFiles {
		fmt.Printf("TODO blogFile: %v\n", blogFile)
	}

	out, err := format.Source(buf.Bytes())
	if err != nil {
		buf.WriteTo(os.Stderr)
		log.Fatal(err)
	}

	if err := ioutil.WriteFile(*outfile, out, 0666); err != nil {
		log.Fatal(err)
	}
}
