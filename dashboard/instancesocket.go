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

func NewInstanceSocket(ws *websocket.Conn, im *client.InstanceMonitor) {
	closeChan := make(chan bool, 1)
	readChan := make(chan SocketRequest)
	ticker := time.NewTicker(5 * time.Second)
	lastHeartbeat := time.Now()

	callback := func(notification client.InstanceListenerNotification) {
		var err error

		switch notification.Type {
		case client.InstanceListenerAddNotification:
			err = websocket.JSON.Send(ws, SocketResponse{Action: "Added", Data: notification.Service})
		case client.InstanceListenerUpdateNotification:
			err = websocket.JSON.Send(ws, SocketResponse{Action: "Updated", Data: notification.Service})
		case client.InstanceListenerRemoveNotification:
			err = websocket.JSON.Send(ws, SocketResponse{Action: "Removed", Data: notification.Service})
		}

		if err != nil {
			closeChan <- true
		}
	}

	go instanceSocketRead(ws, readChan, closeChan)

	l := im.Listen(skynet.UUID(), &client.Query{}, callback)

	err := websocket.JSON.Send(ws, SocketResponse{Action: "List", Data: l.Instances})

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
				err := websocket.JSON.Send(ws, SocketResponse{Action: "list", Data: l.Instances})
				if err != nil {
					closeChan <- true
				}
			case "Heartbeat":
				// this is here more for documentation purposes, setting the lastHeartbeat on read handles the logic here
			}
		}
	}
}
