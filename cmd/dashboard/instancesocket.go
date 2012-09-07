package main

import (
	"code.google.com/p/go.net/websocket"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/client"
	"time"
)

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

	l := im.Listen(skynet.UUID(), &skynet.Query{})

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
				// this is here more for documentation purposes, setting the lastHeartbeat on read handles the logic here
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
