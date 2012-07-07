package skylib

import (
	"bytes"
	"crypto/rand"
	"io"
	"log"
	"fmt"
	"os"
	"time"
	"encoding/json"
	"launchpad.net/mgo/v2"
)

// Logger is an interface to any logging utility.
type Logger interface {
	// Log the item. Can be anything that is representable as a single JSON object.
	Item(item interface{})
	// Log the item for the indicated Service
	ServiceItem(service *Service, item interface{})

	// this function exists only to catch things that are not transitioned
	Println(items ...interface{})
	// this function exists only to catch things that are not transitioned
	Panic(item interface{})
}

func MakeJObj(item interface{}) (jobj map[string]interface{}) {
	jobj = map[string]interface{}{
		"Time":                  time.Now(),
		fmt.Sprintf("%T", item): item,
	}
	return
}

func MakeJObjService(service *Service, item interface{}) (jobj map[string]interface{}) {
	jobj = map[string]interface{}{
		"Time":                  time.Now(),
		"Service":               service.Config.Name,
		fmt.Sprintf("%T", item): item,
	}
	return
}

// MultiLogger is an implementation of the Logger interface thatsends log
// messages out to multiple loggers.
type MultiLogger []Logger

func NewMultiLogger(loggers ...Logger) (ml MultiLogger) {
	ml = loggers
	return
}
func (ml MultiLogger) Item(item interface{}) {
	for _, l := range ml {
		l.Item(item)
	}
}
func (ml MultiLogger) ServiceItem(service *Service, item interface{}) {
	for _, l := range ml {
		l.ServiceItem(service, item)
	}
}
func (ml MultiLogger) Println(items ...interface{}) {
	for _, l := range ml {
		l.Println(items...)
	}
}
func (ml MultiLogger) Panic(item interface{}) {
	for _, l := range ml {
		l.Panic(item)
	}
}

// ConsoleLogger is an implementation of the Logger interface that just
// prints to the console.
type ConsoleLogger struct {
	l              *log.Logger
	untransitioned *log.Logger
}

func NewConsoleLogger(w io.Writer) (cl *ConsoleLogger) {
	cl = &ConsoleLogger{
		l:              log.New(w, "skynet: ", log.LstdFlags),
		untransitioned: log.New(w, "fix-me: ", log.LstdFlags),
	}
	return
}
func (cl *ConsoleLogger) Item(item interface{}) {
	jobj := MakeJObj(item)
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.Encode(jobj)
	cl.l.Print(buf.String())
}
func (cl *ConsoleLogger) ServiceItem(service *Service, item interface{}) {
	jobj := MakeJObjService(service, item)
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.Encode(jobj)
	cl.l.Print(buf.String())
}
func (cl *ConsoleLogger) Println(items ...interface{}) {
	cl.untransitioned.Println(items...)
}
func (cl *ConsoleLogger) Panic(item interface{}) {
	cl.untransitioned.Panic(item)
}

// MongoLogger will archive log items into the specified mongo database.
// It is best of MongoLogger is given a capped collection. However, it
// can do nothing itself to ensure this, with the current state of mgo.
// The collection must be created as capped before any data is inserted.
type MongoLogger struct {
	session         *mgo.Session
	dbName, colName string
	hash            string

	untransitioned *log.Logger
}

func NewMongoLogger(addr string, dbName, collectionName string) (ml *MongoLogger, err error) {
	ml = &MongoLogger{
		dbName:         dbName,
		colName:        collectionName,
		untransitioned: log.New(os.Stderr, "fix-me: ", log.LstdFlags),
		hash:           uuid(),
	}
	ml.session, err = mgo.Dial(addr)
	return
}
func (ml *MongoLogger) Item(item interface{}) {
	jobj := MakeJObj(item)
	jobj["uuid"] = ml.hash

	db := ml.session.DB(ml.dbName)
	col := db.C(ml.colName)

	err := col.Insert(jobj)
	if err != nil {
		log.Printf("Could not log %v: %v", jobj, err)
	}
}
func (ml *MongoLogger) ServiceItem(service *Service, item interface{}) {
	jobj := MakeJObjService(service, item)
	jobj["uuid"] = ml.hash

	db := ml.session.DB(ml.dbName)
	col := db.C(ml.colName)

	err := col.Insert(jobj)
	if err != nil {
		log.Printf("Could not log %v: %v", jobj, err)
	}
}
func (ml *MongoLogger) Println(items ...interface{}) {
	ml.untransitioned.Println(items...)
}
func (ml *MongoLogger) Panic(item interface{}) {
	ml.untransitioned.Panic(item)
}

func uuid() string {
	b := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		log.Fatal(err)
	}
	b[6] = (b[6] & 0x0F) | 0x40
	b[8] = (b[8] &^ 0x40) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:])
}
