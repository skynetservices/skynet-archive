package main

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"
)

import (
	"labix.org/v2/mgo"
)

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
				fmt.Printf("%s: error receiving from client: %s\n", c.ws.Request().RemoteAddr, err)
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
				s := fmt.Sprintf("reader: can not compile regexp: %s %s", message.Filter, err)
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
					s := fmt.Sprintf("%v: %v", k, v) // how complicated are the logs? objects? collections?
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
