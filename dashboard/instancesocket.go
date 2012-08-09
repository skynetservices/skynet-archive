package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/client"
	"time"
)

type SocketResponse struct {
	Action string
	Data   interface{}
}

func NewInstanceSocket(ws *websocket.Conn, im *client.InstanceMonitor) {
	closeChan := make(chan bool, 1)
	readChan := make(chan string)
	ticker := time.NewTicker(5 * time.Second)
	lastHeartbeat := time.Now()

	callback := func(notification client.InstanceListenerNotification) {
		var b []byte

		switch notification.Type {
		case client.InstanceListenerAddNotification:
			b, _ = json.Marshal(SocketResponse{Action: "added", Data: notification.Service})
		case client.InstanceListenerUpdateNotification:
			b, _ = json.Marshal(SocketResponse{Action: "updated", Data: notification.Service})
		case client.InstanceListenerRemoveNotification:
			b, _ = json.Marshal(SocketResponse{Action: "removed", Data: notification.Service})
		}

		if len(b) > 0 {
			_, err := ws.Write(b)

			if err != nil {
				closeChan <- true
			}
		}
	}

	go (func() {
		// Watch for read, if it fails break out of loop and close
		for {
			var data string
			err := websocket.Message.Receive(ws, &data)

			if err != nil {
				closeChan <- true
				break
			}

			readChan <- data
		}
	})()

	l := im.Listen(skynet.UUID(), &client.Query{}, callback)

	b, _ := json.Marshal(SocketResponse{Action: "list", Data: l.Instances})
	ws.Write(b)

	for {
		select {
		case <-closeChan:
			ws.Close()
			l.Close()
		case t := <-ticker.C:
			// Check for timeout
			if t.Sub(lastHeartbeat) > (15 * time.Second) {
				ticker.Stop()
				closeChan <- true
			}
		case data := <-readChan:
      lastHeartbeat = time.Now()

			switch data {
			case "list":
				b, _ := json.Marshal(SocketResponse{Action: "list", Data: l.Instances})
				ws.Write(b)
			case "heartbeat":
        // this is here more for documentation purposes, setting the lastHeartbeat on read handles the logic here
			}
		}
	}
}
