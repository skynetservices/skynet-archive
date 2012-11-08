package main

import (
	"bytes"
	"code.google.com/p/go.net/websocket"
	"flag"
	"fmt"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/client"
	"html"
	"html/template"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

import (
	"labix.org/v2/mgo"
)

var layoutTmpl *template.Template
var indexTmpl *template.Template
var searchTmpl *template.Template

var log skynet.SemanticLogger

func indexHandler(w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	indexTmpl.Execute(buf, r.URL.Path)
	layoutTmpl.Execute(w, template.HTML(buf.String()))
}

var session *mgo.Session

func searchHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	if *debug {
		log.Trace(fmt.Sprintf("%+v", skynet.LogsearchClient{
			r.RemoteAddr, r.Method, r.URL.Path,
		}))
	}

	sdata := make([]string, 0)

	if session == nil {
		session, err = mgo.Dial(*mgoserver)
		if err != nil {
			log.Error(fmt.Sprintf("%+v", skynet.MongoError{
				*mgoserver, "can't connect to server",
			}))
			// Tell client:
			// TODO: proper error pages?
			w.Write([]byte("<html><body>Error establishing MongoDB connection</body></html>"))
			return
		}
		log.Trace(fmt.Sprintf("%+v", skynet.MongoConnected{*mgoserver}))
	}

	var dbs []string
	if *mgodb != "" {
		// Only connect to the supplied database
		dbs = []string{*mgodb}
	} else {
		dbs, err = session.DatabaseNames()
		if err != nil {
			log.Error(fmt.Sprintf("%+v", skynet.MongoError{
				*mgoserver,
				fmt.Sprintf("unable to obtain database list: %s", err),
			}))
			// TODO: proper error pages?
			w.Write([]byte("<html><body>Unable to obtain database list</body></html>"))
			return
		}
	}

	for _, db := range dbs {
		ndb := session.DB(db)
		colls, err := ndb.CollectionNames()
		if err != nil {
			log.Trace(fmt.Sprintf("%+v", skynet.MongoError{
				*mgoserver,
				fmt.Sprintf("unable to obtain collection names: %s", err),
			}))
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

var doozer = flag.String("doozer",
	skynet.GetDefaultEnvVar("SKYNET_DZHOST", skynet.DefaultDoozerdAddr),
	"initial doozer instance to connect to")
var doozerboot = flag.String("doozerboot",
	skynet.GetDefaultEnvVar("SKYNET_DZNSHOST", ""),
	"initial doozer instance to connect to")
var autodiscover = flag.Bool("autodiscover",
	skynet.GetDefaultEnvVar("SKYNET_DZDISCOVER", "true") == "true",
	"auto discover new doozer instances")

var debug = flag.Bool("d", false, "print debug info")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var webroot = flag.String("webroot", ".",
	"root of templates and javascript libraries")
var mgoserver = flag.String("mgoserver",
	skynet.GetDefaultEnvVar("SKYNET_MGOSERVER", ""),
	"comma-separated list of urls of mongodb servers")
var mgodb = flag.String("mgodb",
	skynet.GetDefaultEnvVar("SKYNET_MGODB", ""),
	"mongodb database")

var DC *skynet.DoozerConnection

func main() {
	var err error

	flag.Parse()

	log = skynet.NewConsoleSemanticLogger("dashboard", os.Stderr)
	if *mgoserver == "" {
		log.Trace(fmt.Sprintf("%+v", skynet.MongoError{
			"",
			"No mongodb server url (both -mgoserver and SKYNET_MGOSERVER missing)",
		}))
	}

	mlogger, err := skynet.NewMongoSemanticLogger(*mgoserver, "skynet",
		"log", skynet.UUID())
	if err != nil {
		log.Error(fmt.Sprintf("%+v", skynet.MongoError{
			Addr: "Could not connect to mongo db for logging",
			Err: err.Error(),
		}))
	}
	log = skynet.NewMultiSemanticLogger(mlogger, log)

	DC = Doozer()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/logs/search", searchHandler)
	http.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir(*webroot+"/tmpl"))))
	http.Handle("/favicon.ico", http.FileServer(http.Dir(*webroot+"/tmpl/images")))
	http.Handle("/logs/ws", websocket.Handler(wsHandler))

	im := client.NewInstanceMonitor(DC, true)
	http.Handle("/instances/ws", websocket.Handler(func(ws *websocket.Conn) {
		NewInstanceSocket(ws, im)
	}))

	// Cache templates
	layoutTmpl = template.Must(template.ParseFiles(*webroot +
		"/tmpl/layout.html.template"))
	indexTmpl = template.Must(template.ParseFiles(*webroot +
		"/tmpl/index.html.template"))
	searchTmpl = template.Must(template.ParseFiles(*webroot +
		"/tmpl/search.html.template"))

	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: " + err.Error())
	}
}

func Doozer() *skynet.DoozerConnection {
	defer func() {
		if r := recover(); r != nil {
			log.Error("Failed to connect to Doozer")
			os.Exit(1)
		}
	}()

	// nil as the last param will default to a Stdout logger
	conn := skynet.NewDoozerConnection(*doozer, *doozerboot, true, nil)
	conn.Connect()

	return conn
}

//
// Beginning of what was logstreamer.go
//

type connection struct {
	ws     *websocket.Conn
	db     string // mongodb database currently in use
	coll   string // mongodb collection currently tailing
	cancel chan bool
	sess   *mgo.Session
	filter *regexp.Regexp // if set, the client only sees matches
}

type Req struct {
	Collection string `json:"collection"`
	Filter     string `json:"filter"`
}

func (c *connection) send(s string) bool {
	err := websocket.Message.Send(c.ws, s)
	if err != nil {
		c.ws.Close()
		return false
	}
	return true
}

func (c *connection) fromClient() {
	shouldCancel := false
	message := &Req{}
	for {
		err := websocket.JSON.Receive(c.ws, message)
		if err != nil {
			if *debug {
				fmt.Printf("%s: error receiving from client: %s\n",
					c.ws.Request().RemoteAddr, err)
			}
			break
		}
		if *debug {
			fmt.Printf("%s: fromClient: %+v\n", c.ws.Request().RemoteAddr, message)
		}

		if shouldCancel {
			c.cancel <- true
			shouldCancel = false
		}

		if message.Filter != "" {
			c.filter, err = regexp.Compile(html.UnescapeString(message.Filter))
			if err != nil {
				s := fmt.Sprintf("reader: can not compile regexp: %s %s",
					message.Filter, err)
				if !c.send(s) {
					return
				}
				if *debug {
					fmt.Printf("%s: %s\n", c.ws.Request().RemoteAddr, s)
				}
				continue
			}
		} else {
			c.filter = nil
		}

		if message.Collection != "" {
			dbc := strings.Split(message.Collection, ":")
			if len(dbc) != 2 {
				s := fmt.Sprintf("internal error: received bad db:collection from client: %s", message.Collection)
				if !c.send(s) {
					return
				}
				if *debug {
					fmt.Printf("%s: %s\n", c.ws.Request().RemoteAddr, s)
				}
				continue
			}
			c.db = dbc[0]
			c.coll = dbc[1]
		} else {
			s := fmt.Sprintf("internal error: db:collection shouldn't be nil")
			if !c.send(s) {
				return
			}
			if *debug {
				fmt.Printf("%s: %s\n", c.ws.Request().RemoteAddr, s)
			}
			c.db = ""
			c.coll = ""
			continue
		}

		if c.coll != "" {
			go c.dump()
			shouldCancel = true
		}
	}
	c.ws.Close()
}

func (c *connection) dump() {
	var result = make(map[string]interface{})

	coll := c.sess.DB(c.db).C(c.coll)
	if coll == nil {
		if !c.send("internal mgo error: shouldn't happen!") {
			return
		}
		<-c.cancel
		return
	}
	iter := coll.Find(nil).Tail(500 * time.Millisecond)
	if iter.Err() != nil {
		s := fmt.Sprintf("internal error: %s", iter.Err())
		if !c.send(s) {
			return
		}
		if *debug {
			fmt.Printf("%s: %s\n", c.ws.Request().RemoteAddr, s)
		}
		// we must block here, no need to continue spinning
		<-c.cancel
		return
	}

	// Need to spin to be able to consume a cancel
	for {
		select {
		case <-c.cancel:
			return
		default:
			if iter.Next(result) {
				for k, v := range result {
					// how complicated are the logs? objects? collections?
					s := fmt.Sprintf("%v: %v", k, v)
					if c.filter == nil || c.filter.MatchString(s) {
						if !c.send(s) {
							return
						}
					}
				}
				if iter.Err() != nil {
					s := fmt.Sprintf("%s", iter.Err())
					if !c.send(s) {
						return
					}
					<-c.cancel
					return
				}
			} else {
				if iter.Err() != nil {
					s := fmt.Sprintf("%s", iter.Err())
					if !c.send(s) {
						return
					}
					<-c.cancel
					return
				}
				if !iter.Timeout() {
					if !c.send("lost connection to server, won't retry") {
						return
					}
					<-c.cancel
					return
				}
			}
		}
	}
}

func wsHandler(ws *websocket.Conn) {
	// Would it be better for each individual client to open 
	// a separate connection with the MongoDB server by calling Dial here?
	c := &connection{ws: ws, cancel: make(chan bool), sess: session}
	c.fromClient() // must wait for client to select database
	ws.Close()
	close(c.cancel) // ensure no dangling readers
}

//
// Beginning of what was instancesocket.go
//

type SocketResponse struct {
	Action string
	Data   interface{}
}

type SocketRequest struct {
	Action string
	Data   interface{}
}

func instanceSocketRead(ws *websocket.Conn, readChan chan SocketRequest, closeChan chan bool) {
	// Watch for read, if it fails break out of loop and close
	for {
		var request SocketRequest
		err := websocket.JSON.Receive(ws, &request)

		if err != nil {
			closeChan <- true
			break
		}

		readChan <- request
	}
}

func sendInstanceList(ws *websocket.Conn, im *client.InstanceMonitor) {
	// Wait for list to be built, then pull off the notification channel
}

func NewInstanceSocket(ws *websocket.Conn, im *client.InstanceMonitor) {
	closeChan := make(chan bool, 1)
	readChan := make(chan SocketRequest)
	ticker := time.NewTicker(5 * time.Second)
	lastHeartbeat := time.Now()

	go instanceSocketRead(ws, readChan, closeChan)

	l := im.Listen(skynet.UUID(), &skynet.Query{}, true)

	instances := <-l.NotificationChan
	err := websocket.JSON.Send(ws, SocketResponse{Action: "List", Data: instances})

	if err != nil {
		closeChan <- true
	}

	for {
		select {
		case <-closeChan:
			ticker.Stop()
			ws.Close()
			l.Close()
		case t := <-ticker.C:
			// Check for timeout
			if t.Sub(lastHeartbeat) > (15 * time.Second) {
				closeChan <- true
			}
		case request := <-readChan:
			lastHeartbeat = time.Now()

			switch request.Action {
			case "List":
				err := websocket.JSON.Send(ws, SocketResponse{Action: "List", Data: instances})
				if err != nil {
					closeChan <- true
				}
			case "Heartbeat":
				// this is here more for documentation purposes,
				// setting the lastHeartbeat on read handles the logic
				// here
			case "Filter":
				if request.Data != nil {
					data := request.Data.(map[string]interface{})

					if r, ok := data["Reset"]; ok {
						reset := r.(bool)
						if reset {
							l.Query.Reset()
						}
					}

					if r, ok := data["Registered"]; ok {
						filter := r.(bool)
						l.Query.Registered = &filter
					}
				}

				instances := l.GetInstances()
				iln := make(client.InstanceListenerNotification)
				for _, i := range instances {
					path := i.GetConfigPath()
					iln[path] = client.InstanceMonitorNotification{
						Path:    path,
						Service: i,
						Type:    client.InstanceAddNotification,
					}
				}

				err := websocket.JSON.Send(ws, SocketResponse{Action: "List", Data: iln})

				if err != nil {
					closeChan <- true
				}
			}

		case notification := <-l.NotificationChan:
			var err error

			// Forward message as it stands across the websocket
			err = websocket.JSON.Send(ws, SocketResponse{Action: "Update", Data: notification})

			instances = instances.Join(notification)

			if err != nil {
				closeChan <- true
			}
		}
	}
}
