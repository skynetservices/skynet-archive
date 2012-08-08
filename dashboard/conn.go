package main

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"html"
	"log"
	"regexp"
)

type connection struct {
	ws     *websocket.Conn
	send   chan string
	filter *regexp.Regexp // if set, the client only sees matches
}

func (c *connection) reader() {
	for {
		var message string
		err := websocket.Message.Receive(c.ws, &message)
		if err != nil {
			log.Printf("reader: bad read: %s\n", err)
			break
		}
		if *debug {
			log.Printf("reader received: %s", message)
		}

		// Algorithm is as follows: unregister the client from the broadcast
		// queue, then try to compile the regex. Then call logdump() to grab all 
		// data from the log and then register back to receive future matching log
		// messages 
		h.unregister <- c

		str := html.UnescapeString(message)
		c.filter, err = regexp.Compile(html.UnescapeString(message))
		if err != nil {
			s := fmt.Sprintf("reader: can not compile regexp: %s %s\n", str, err)
			log.Print(s)
			err := websocket.Message.Send(c.ws, s)
			if err != nil {
				break
			}
		} else {
			dump(c.send)
			h.register <- c
		}
	}
	c.ws.Close()
}

func (c *connection) writer() {
	for message := range c.send {
		if *debug {
			log.Printf("writer received: %s", message)
		}
		if c.filter == nil || c.filter.MatchString(message) {
			err := websocket.Message.Send(c.ws, message)
			if err != nil {
				break
			}
		}
	}
	c.ws.Close()
}

func wsHandler(ws *websocket.Conn) {
	c := &connection{send: make(chan string, 256), ws: ws, filter: nil}
	h.register <- c
	defer func() { h.unregister <- c }()
	go c.reader()
	c.writer()
}
