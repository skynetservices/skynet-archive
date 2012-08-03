package main

import (
	"code.google.com/p/go.net/websocket"
	"flag"
	"html/template"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"
  "bytes"
)

var layoutTmpl *template.Template
var indexTmpl *template.Template
var searchTmpl *template.Template

func indexHandler(w http.ResponseWriter, r *http.Request) {
  buf := new(bytes.Buffer)
	indexTmpl.Execute(buf, r.URL.Path)
  layoutTmpl.Execute(w, template.HTML(buf.String()))
}

func searchHandler(w http.ResponseWriter, req *http.Request) {
  buf := new(bytes.Buffer)
	searchTmpl.Execute(buf, req.Host)
  layoutTmpl.Execute(w, template.HTML(buf.String()))
}

var addr = flag.String("addr", ":8080", "dashboard listener address")
var debug = flag.Bool("d", false, "print debug info")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var webroot = flag.String("webroot", ".", "root of templates and javascript libraries")
var mems = flag.Bool("memstats", false, "write mem stats to stderr")
var memstats *runtime.MemStats

func main() {
	flag.Parse()

	if *mems {
		memstats = new(runtime.MemStats)
	}
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	if *mems {
		runtime.ReadMemStats(memstats)
		log.Printf("memstats GC: bytes = %d footprint = %d\n", memstats.HeapAlloc, memstats.Sys)
		log.Printf("memstats GC: %v\n", memstats.PauseNs)
	}
	if *debug {
		log.Printf("preparing web server splash page...\n")
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/logs/search", searchHandler)
	http.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir(*webroot+"/tmpl"))))
	http.Handle("/favicon.ico", http.FileServer(http.Dir(*webroot+"/tmpl/images")))
	http.Handle("/logs/ws", websocket.Handler(wsHandler))

	// Cache templates
	layoutTmpl = template.Must(template.ParseFiles(*webroot + "/tmpl/layout.html.template"))
	indexTmpl = template.Must(template.ParseFiles(*webroot + "/tmpl/index.html.template"))
	searchTmpl = template.Must(template.ParseFiles(*webroot + "/tmpl/search.html.template"))

	// Start logging service hub
	go h.run()
	// Start dummy log fetcher
	go logbroadcast()

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
