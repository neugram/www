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

var tmpl = template.Must(template.New("blogentry").Parse(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>The Neugram Blog: {{.Title}}</title>
<meta name=viewport content="width=device-width, initial-scale=1"/>
<style>
{{template "style.css"}}
</style>
</head>
<body>
{{.Contents}}

<a href="/">Home</a>,
<a href="/blog/">Blog Archive</a>

{{template "footer"}}
</body>
</html>
`))

var footer = template.Must(tmpl.New("footer").Parse(`<script>
window.ga=window.ga||function(){(ga.q=ga.q||[]).push(arguments)};ga.l=+new Date;
ga('create', 'UA-92251090-1', 'auto');
ga('send', 'pageview');
</script>
<script async src='https://www.google-analytics.com/analytics.js'></script>
`))

var bloglist = template.Must(tmpl.New("bloglist").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>The Neugram Blog</title>
<meta name=viewport content="width=device-width, initial-scale=1"/>
<link href="https://neugram.io/atom.xml" rel="alternate" type="application/rss+xml" title="The Neugram Blog" />
<style>
{{template "style.css"}}

.entries div { display: flex; }
.entrydate   { width: 6em; text-align; right; flex-shrink: 0; }
</style>
</head>
<body>
<h1>The Neugram Blog</h1>
<div class="entries">
{{range .}}
<div><div class="entrydate">{{.Date}}</div><a href="{{.URL}}">{{.Title}}</a></div>
{{end}}
</div>

{{template "footer"}}
</body>
</html>
`))

var blogatom = template.Must(tmpl.New("bloglist").Parse(`
<rss xmlns:atom="http://www.w3.org/2005/Atom" version="2.0">
<channel>
<title>The Neugram Blog</title>
<link>https://neugram.io/atom.xml</link>
<description>Recent content on The Neugram Blog</description>
<generator>custom neugram</generator>
<language>en-us</language>
<lastBuildDate>{{.LatestDate}}</lastBuildDate>
<atom:link href="https://neugram.io/atom.xml" rel="self" type="application/rss+xml"/>

{{range .Entries}}
<item>
<title>{{.Title}}</title>
<link>https://neugram.io{{.URL}}</link>
<pubDate>{{.PubDate}}</pubDate>
<guid>https://neugram.io{{.URL}}</guid>
<description>
{{.ContentsQuoted}}
</description>
</item>
{{end}}

</channel>
</rss>
`))

type BlogEntry struct {
	URL            string
	Title          string
	Contents       template.HTML
	ContentsQuoted string
	Date           string // for humans
	PubDate        string // for machines: "Mon, 2 Jan 2006 00:00:00 +0000"
}

func writeFile(buf *bytes.Buffer, name string, path string) {
	orig, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	var src []byte
	if filepath.Base(path) == "main.html" {
		// TODO: apply to all .html files
		// Treat this as a template.
		t := template.Must(tmpl.New(name).Parse(string(orig)))
		srcbuf := new(bytes.Buffer)
		if err := t.Execute(srcbuf, nil); err != nil {
			log.Fatal(err)
		}
		src = srcbuf.Bytes()
	} else {
		src = orig
	}
	writeBytes(buf, name, src)
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
		if filepath.Ext(blogFile) != ".md" {
			writeFile(buf, "/"+blogFile, blogFile)
			continue
		}
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
		if err = tmpl.Execute(srcbuf, entry); err != nil {
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
}

func main() {
	flag.Parse()

	cssb, err := ioutil.ReadFile("style.css")
	if err != nil {
		log.Fatal(err)
	}
	template.Must(tmpl.New("style.css").Parse(string(cssb)))

	buf := new(bytes.Buffer)
	fmt.Fprint(buf, "package main\n\n")

	writeFile(buf, "/", "main.html")
	writeFile(buf, "/ng/", "ng.html")
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
