package main

import (
	"bytes"
	"code.google.com/p/go.net/websocket"
	"flag"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/client"
	"github.com/skynetservices/skynet/log"
	"html/template"
	"net/http"
	"os"
	"time"
)

var layoutTmpl *template.Template
var indexTmpl *template.Template

func indexHandler(w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	indexTmpl.Execute(buf, r.URL.Path)
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

var DC *skynet.DoozerConnection

func main() {
	var err error

	flag.Parse()

	DC = Doozer()

	http.HandleFunc("/", indexHandler)
	http.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir(*webroot+"/tmpl"))))
	http.Handle("/favicon.ico", http.FileServer(http.Dir(*webroot+"/tmpl/images")))

	im := client.NewInstanceMonitor(DC, true)
	http.Handle("/instances/ws", websocket.Handler(func(ws *websocket.Conn) {
		NewInstanceSocket(ws, im)
	}))

	// Cache templates
	layoutTmpl = template.Must(template.ParseFiles(*webroot +
		"/tmpl/layout.html.template"))
	indexTmpl = template.Must(template.ParseFiles(*webroot +
		"/tmpl/index.html.template"))

	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: " + err.Error())
	}
}

func Doozer() *skynet.DoozerConnection {
	defer func() {
		if r := recover(); r != nil {
			log.Panic("Failed to connect to Doozer")
			os.Exit(1)
		}
	}()

	conn := skynet.NewDoozerConnection(*doozer, *doozerboot, true)
	conn.Connect()

	return conn
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
