package main

import (
	"bytes"
	"code.google.com/p/go.net/websocket"
	"flag"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/client"
	"html/template"
	"net/http"
	"fmt"
	"os"
)

import (
	"labix.org/v2/mgo"
)

var layoutTmpl *template.Template
var indexTmpl *template.Template
var searchTmpl *template.Template

var log skynet.Logger

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if *debug {
		log.Item(fmt.Sprintf("%s → %s %s", r.RemoteAddr, r.Method, r.URL.Path))
	}
	buf := new(bytes.Buffer)
	indexTmpl.Execute(buf, r.URL.Path)
	layoutTmpl.Execute(w, template.HTML(buf.String()))
}

var session *mgo.Session

func searchHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	if *debug {
		log.Item(fmt.Sprintf("%s → %s %s", r.RemoteAddr, r.Method, r.URL.Path))
	}

	sdata := make([]string, 0)

	if session == nil {
		session, err = mgo.Dial(*mgoserver)
		if err != nil {
			log.Item(fmt.Sprintf("searchHandler: can't connect to mongodb server %s: %s\n", *mgoserver, err))
			// TODO: proper error pages?
			w.Write([]byte("<html><body>Error establishing MongoDB connection</body></html>"))
			return
		}
	}

	var dbs []string
	if *mgodb != "" {
		// Only connect to the supplied database
		dbs = []string{*mgodb}
	} else {
		dbs, err = session.DatabaseNames()
		if err != nil {
			log.Println("searchHandler: unable to obtain database list: %s\n", err)
			// TODO: proper error pages?
			w.Write([]byte("<html><body>Unable to obtain database list</body></html>"))
			return
		}
	}

	for _, db := range dbs {
		ndb := session.DB(db)
		colls, err := ndb.CollectionNames()
		if err != nil {
			log.Item(fmt.Sprintf("searchHandler: can't get collection names for %s: %s", db, err))
			continue
		}
		for _, coll := range colls {
			sdata = append(sdata, db+":"+coll)
		}
	}

	buf := new(bytes.Buffer)
	searchTmpl.Execute(buf, sdata)
	layoutTmpl.Execute(w, template.HTML(buf.String()))
}

var addr = flag.String("addr", ":8080", "dashboard listener address")

var doozer = flag.String("doozer", skynet.GetDefaultEnvVar("DZHOST", "127.0.0.1:8046"), "initial doozer instance to connect to")
var doozerboot = flag.String("doozerboot", skynet.GetDefaultEnvVar("DZNSHOST", ""), "initial doozer instance to connect to")
var autodiscover = flag.Bool("autodiscover", skynet.GetDefaultEnvVar("DZDISCOVER", "true") == "true", "auto discover new doozer instances")

var debug = flag.Bool("d", false, "print debug info")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var webroot = flag.String("webroot", ".", "root of templates and javascript libraries")
var mgoserver = flag.String("mgoserver", skynet.GetDefaultEnvVar("SKYNET_MGOSERVER", ""), "comma-separated list of urls of mongodb servers")
var mgodb = flag.String("mgodb", skynet.GetDefaultEnvVar("SKYNET_MGODB", ""), "mongodb database")

var DC *skynet.DoozerConnection

func main() {
	var err error

	flag.Parse()

	log = skynet.NewConsoleLogger(os.Stderr)

	if *mgoserver == "" {
		log.Panic("no mongodb server url (both -mgoserver and SKYNET_MGOSERVER missing)")
	}

	DC = Doozer()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/logs/search", searchHandler)
	http.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir(*webroot+"/tmpl"))))
	http.Handle("/favicon.ico", http.FileServer(http.Dir(*webroot+"/tmpl/images")))
	http.Handle("/logs/ws", websocket.Handler(wsHandler))

	im := client.NewInstanceMonitor(DC)
	http.Handle("/instances/ws", websocket.Handler(func(ws *websocket.Conn) {
		NewInstanceSocket(ws, im)
	}))

	// Cache templates
	layoutTmpl = template.Must(template.ParseFiles(*webroot + "/tmpl/layout.html.template"))
	indexTmpl = template.Must(template.ParseFiles(*webroot + "/tmpl/index.html.template"))
	searchTmpl = template.Must(template.ParseFiles(*webroot + "/tmpl/search.html.template"))

	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Panic("ListenAndServe: " + err.Error())
	}
}

func Doozer() *skynet.DoozerConnection {
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
