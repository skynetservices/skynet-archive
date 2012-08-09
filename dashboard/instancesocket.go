package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/client"
)

type SocketResponse struct {
	Action string
	Data   interface{}
}

func NewInstanceSocket(ws *websocket.Conn, im *client.InstanceMonitor) {
	l := im.Listen(skynet.UUID(), &client.Query{})

	b, _ := json.Marshal(SocketResponse{Action: "list", Data: l.Instances})
	ws.Write(b)

	// TODO: make sure this goes out of scope when the user closes the socket or times out (send heartbeat?)
	// Close the websocket, and remove the listener from the InstanceMonitor: l.Close()
	for {
		select {
    case notification := <-l.NotificationChan:
      var b []byte
      switch notification.Type {
        case client.InstanceListenerAddNotification:
          b, _ = json.Marshal(SocketResponse{Action: "added", Data: notification.Service})
        case client.InstanceListenerUpdateNotification:
          b, _ = json.Marshal(SocketResponse{Action: "updated", Data: notification.Service})
        case client.InstanceListenerRemoveNotification:
          b, _ = json.Marshal(SocketResponse{Action: "removed", Data: notification.Service})
      }

			ws.Write(b)
		}
	}
}
