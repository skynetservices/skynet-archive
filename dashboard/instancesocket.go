package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/client"
)

func NewInstanceSocket(ws *websocket.Conn, im *client.InstanceMonitor) {
	l := im.Listen(skynet.UUID(), &client.Query{})

	b, _ := json.Marshal(l.Instances)
	ws.Write(b)

	// TODO: make sure this goes out of scope when the user closes the socket or times out (send heartbeat?)
	// Close the websocket, and remove the listener from the InstanceMonitor: l.Close()
	for {
		select {
		case service := <-l.AddChan:
			b, _ := json.Marshal(service)

			ws.Write(b)
		case path := <-l.RemoveChan:
			ws.Write([]byte(path))
		}
	}
}
