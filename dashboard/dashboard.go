package main

import (
	"bytes"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/client"
	"code.google.com/p/go.net/websocket"
	"flag"
	"html/template"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"
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
	if *debug {
		log.Printf("%s → %s %s", req.RemoteAddr, req.Method, req.URL.Path)
	}
	buf := new(bytes.Buffer)
	searchTmpl.Execute(buf, req.Host)
	layoutTmpl.Execute(w, template.HTML(buf.String()))
}

var addr = flag.String("addr", ":8080", "dashboard listener address")

var doozer = flag.String("doozer", skynet.GetDefaultEnvVar("DZHOST", "127.0.0.1:8046"), "initial doozer instance to connect to")
var doozerboot = flag.String("doozerboot", skynet.GetDefaultEnvVar("DZNSHOST", ""), "initial doozer instance to connect to")
var autodiscover = flag.Bool("autodiscover", skynet.GetDefaultEnvVar("DZDISCOVER", "true") == "true", "auto discover new doozer instances")

var debug = flag.Bool("d", false, "print debug info")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var webroot = flag.String("webroot", ".", "root of templates and javascript libraries")
var mems = flag.Bool("memstats", false, "write mem stats to stderr")
var memstats *runtime.MemStats

var DC skynet.DoozerConnection

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

  DC = Doozer() 

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/logs/search", searchHandler)
	http.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir(*webroot+"/tmpl"))))
	http.Handle("/favicon.ico", http.FileServer(http.Dir(*webroot+"/tmpl/images")))
	http.Handle("/logs/ws", websocket.Handler(wsHandler))

  im := client.NewInstanceMonitor(DC)

	http.Handle("/instances/ws", websocket.Handler(func (ws *websocket.Conn){
    NewInstanceSocket(ws, im)
  }))

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
func Doozer() skynet.DoozerConnection {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Failed to connect to Doozer")
			os.Exit(1)
		}
	}()

	conn := skynet.NewDoozerConnection(*doozer, *doozerboot, true, nil) // nil as the last param will default to a Stdout logger
	conn.Connect()

	return conn
}
