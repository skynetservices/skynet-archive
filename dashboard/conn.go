package main

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"html"
	"io"
	"log"
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

func (c *connection) fromClient() {
	shouldCancel := false
	message := &Req{}
	for {
		err := websocket.JSON.Receive(c.ws, message)
		if err != nil {
			if err != io.EOF {
				log.Printf("fromClient: bad receive: %s", err)
			}
			break
		}
		if *debug {
			log.Printf("fromClient: %+v", message)
		}

		if message.Filter != "" || message.Collection != "" {
			// make sure we've cancelled before we assign new values to the object
			if shouldCancel {
				c.cancel <- true
				shouldCancel = false
			}
		}

		if message.Filter != "" {
			c.filter, err = regexp.Compile(html.UnescapeString(message.Filter))
			if err != nil {
				s := fmt.Sprintf("reader: can not compile regexp: %s %s\n", message.Filter, err)
				log.Printf("%s", s)
				websocket.Message.Send(c.ws, s)
				continue
			}
		}

		if message.Collection != "" {
			dbc := strings.Split(message.Collection, ":")
			if len(dbc) != 2 {
				s := fmt.Sprintf("internal error: received bad db:collection from client: %s", message.Collection)
				log.Printf("%s", s)
				websocket.Message.Send(c.ws, s)
				continue
			}
			c.db = dbc[0]
			c.coll = dbc[1]
		}
		if c.coll != "" {
			go c.dump()
			shouldCancel = true
		}
	}
	c.ws.Close()
}

func (c *connection) dump() {
	var iter *mgo.Iter
	var result = make(map[string]interface{})

	coll := c.sess.DB(c.db).C(c.coll)
	query := coll.Find(nil)
	iter = query.Iter()

	// Need to spin to be able to consume a cancel
outer:
	for {
		select {
		case <-c.cancel:
			return
		default:
			if iter.Next(result) {
				for k, v := range result {
					s := fmt.Sprintf("%v: %v", k, v) // how complicated are the logs? objects? collections?
					if c.filter == nil || c.filter.MatchString(s) {
						websocket.Message.Send(c.ws, s)
					}
				}
			} else {
				if iter.Err() != nil {
					s := fmt.Sprintf("internal error: %s", iter.Err())
					log.Printf("%s", s)
					websocket.Message.Send(c.ws, s)
					// we must block here, no need to continue spinning
					<-c.cancel
				}
				break outer
			}
		}
	}

	// now tail for updated results
	iter = query.Tail(500 * time.Millisecond)
	for {
		select {
		case <-c.cancel:
			break
		default:
			if iter.Next(result) {
				for k, v := range result {
					s := fmt.Sprintf("%v: %v", k, v) // how complicated are the logs? objects? collections?
					if c.filter == nil || c.filter.MatchString(s) {
						websocket.Message.Send(c.ws, s)
					}
				}
			}
			if !iter.Timeout() {
				log.Println("iter is done", iter.Err().Error())
				<-c.cancel
			}
		}
	}
}

func wsHandler(ws *websocket.Conn) {
	// Would it be better for each individual client to open 
	// a separate connection with the MongoDB server by calling Dial here?
	c := &connection{ws: ws, cancel: make(chan bool), sess: session}

	c.fromClient() // must wait for client to select database
}
