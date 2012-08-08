package main

import (
	"bytes"
	"code.google.com/p/go.net/websocket"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
)

import (
	"labix.org/v2/mgo"
)

var layoutTmpl *template.Template
var indexTmpl *template.Template
var searchTmpl *template.Template

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if *debug {
		log.Printf("%s → %s %s", r.RemoteAddr, r.Method, r.URL.Path)
	}
	buf := new(bytes.Buffer)
	indexTmpl.Execute(buf, r.URL.Path)
	layoutTmpl.Execute(w, template.HTML(buf.String()))
}

var session *mgo.Session

func searchHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	if *debug {
		log.Printf("%s → %s %s", r.RemoteAddr, r.Method, r.URL.Path)
	}

	sdata := make([]string, 0)

	if session == nil {
		session, err = mgo.Dial(*mgoserver)
		if err != nil {
			log.Printf("searchHandler: can't connect to mongodb server %s: %s\n", *mgoserver, err)
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
			log.Printf("searchHandler: unable to obtain database list: %s\n", err)
			// TODO: proper error pages?
			w.Write([]byte("<html><body>Unable to obtain database list</body></html>"))
			return
		}
	}

	for _, db := range dbs {
		ndb := session.DB(db)
		colls, err := ndb.CollectionNames()
		if err != nil {
			log.Printf("searchHandler: can't get collection names for %s: %s", db, err)
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
var debug = flag.Bool("d", false, "print debug info")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var webroot = flag.String("webroot", ".", "root of templates and javascript libraries")
var mgoserver = flag.String("mgoserver", "", "comma-separated list of urls of mongodb servers")
var mgodb = flag.String("mgodb", "", "mongodb database")

func main() {
	var err error

	flag.Parse()

	if *mgoserver == "" {
		if *mgoserver = os.Getenv("SKYNET_MGOSERVER"); *mgoserver == "" {
			log.Fatal("no mongodb server url (both -mgoserver and SKYNET_MGOSERVER missing)")
		}
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

	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
