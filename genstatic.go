// +build ignore

// TODO: rewrite this in Neugram!

package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"flag"
	"fmt"
	"go/format"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	blackfriday "gopkg.in/russross/blackfriday.v2"
)

var outfile = flag.String("o", "", "result will be written file")

var neusvg = loadTmpl("neu.svg")
var bloglist = loadTmpl("blog.html")
var blogatom = loadTmpl("atom.xml")
var blogentry = loadTmpl("blogentry.html")
var mainhtml = loadTmpl("main.html")
var nghtml = loadTmpl("ng.html")
var stylecss = loadTmpl("style.css")
var footer = template.Must(template.New("footer").Parse(`<script>
window.ga=window.ga||function(){(ga.q=ga.q||[]).push(arguments)};ga.l=+new Date;
ga('create', 'UA-92251090-1', 'auto');
ga('send', 'pageview');
</script>
<script async src='https://www.google-analytics.com/analytics.js'></script>
`))

func loadTmpl(filename string) (t *template.Template) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	return template.Must(footer.New(filename).Parse(string(b)))
}

type BlogEntry struct {
	URL            string
	Title          string
	Contents       template.HTML
	ContentsQuoted string
	Date           string // for humans
	PubDate        string // for machines: "Mon, 2 Jan 2006 00:00:00 +0000"
}

func writeBytes(buf *bytes.Buffer, name string, src []byte) {
	log.Printf("writing %s", name)
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
	fmt.Fprintf(buf, "\t``\n}\n\n")
}

var mdNameRE = regexp.MustCompile(`(\d\d\d\d-\d\d-\d\d)-(.*).md`)
var titleRE = regexp.MustCompile(`^# (.*)\n`)

func writeBlogFiles(buf *bytes.Buffer) {
	blogFiles, err := filepath.Glob("blog/*.*")
	if err != nil {
		log.Fatal(err)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(blogFiles)))
	var entries []BlogEntry
	for _, blogFile := range blogFiles {
		src, err := ioutil.ReadFile(blogFile)
		if err != nil {
			log.Fatal(err)
		}
		match := mdNameRE.FindStringSubmatch(filepath.Base(blogFile))
		date, urlTitle := match[1], match[2]
		dateVal, err := time.Parse("2006-01-02", date)
		if err != nil {
			log.Fatalf("%q date: %v", blogFile, err)
		}

		titleMatch := titleRE.FindSubmatch(src)
		if titleMatch == nil {
			log.Fatalf("%s: no title found", blogFile)
		}
		title := string(titleMatch[1])

		url := "/blog/" + urlTitle
		out := blackfriday.Run(src)
		srcbuf := new(bytes.Buffer)
		entry := BlogEntry{
			URL:            url,
			Title:          title,
			Date:           date,
			PubDate:        dateVal.Format("Mon, 2 Jan 2006 00:00:00 +0000"),
			Contents:       template.HTML(out),
			ContentsQuoted: string(out),
		}
		if err = blogentry.Execute(srcbuf, entry); err != nil {
			log.Fatal(err)
		}
		entries = append(entries, entry)
		writeBytes(buf, url, srcbuf.Bytes())
	}

	srcbuf := new(bytes.Buffer)
	if err := bloglist.Execute(srcbuf, entries); err != nil {
		log.Fatal(err)
	}
	writeBytes(buf, "/blog/", srcbuf.Bytes())

	type RSS struct {
		LatestDate string
		Entries    []BlogEntry
	}
	rss := RSS{
		LatestDate: entries[0].PubDate,
		Entries:    entries,
	}
	srcbuf = new(bytes.Buffer)
	if err := blogatom.Execute(srcbuf, rss); err != nil {
		log.Fatal(err)
	}
	writeBytes(buf, "/atom.xml", srcbuf.Bytes())

	if len(entries) > 5 {
		entries = entries[:5]
	}
	srcbuf = new(bytes.Buffer)
	if err := mainhtml.Execute(srcbuf, entries); err != nil {
		log.Fatal(err)
	}
	writeBytes(buf, "/", srcbuf.Bytes())
}

func main() {
	flag.Parse()

	buf := new(bytes.Buffer)
	fmt.Fprint(buf, "package main\n\n")

	srcbuf := new(bytes.Buffer)
	if err := nghtml.Execute(srcbuf, nil); err != nil {
		log.Fatal(err)
	}
	writeBytes(buf, "/ng/", srcbuf.Bytes())

	b, err := ioutil.ReadFile("favicon.png")
	if err != nil {
		log.Fatal(err)
	}
	writeBytes(buf, "/favicon.png", b)

	writeBlogFiles(buf)

	out, err := format.Source(buf.Bytes())
	if err != nil {
		buf.WriteTo(os.Stderr)
		log.Fatal(err)
	}

	if err := ioutil.WriteFile(*outfile, out, 0666); err != nil {
		log.Fatal(err)
	}
}
