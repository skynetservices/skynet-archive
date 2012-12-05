package skynet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/skynetservices/mgo"
	"io"
	"log"
	"os"
	"runtime"
	"time"
)

// Logger is an interface to any logging utility.
type Logger interface {
	// Log the item. Can be anything that is representable as a single JSON object.
	Item(item interface{})
	// Something has gone horribly wrong - remember what and bomb the program.
	Panic(item interface{})
	// these functions exists only to catch things that are not transitioned
	Println(items ...interface{})
}

func MakeJObj(item interface{}) (jobj map[string]interface{}) {
	jobj = map[string]interface{}{
		"Time":                  time.Now(),
		fmt.Sprintf("%T", item): item,
	}
	return
}

// MultiLogger is an implementation of the Logger interface that sends log
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

func NewConsoleLogger(name string, w io.Writer) (cl *ConsoleLogger) {
	cl = &ConsoleLogger{
		l:              log.New(w, fmt.Sprintf("%s: ", name), log.LstdFlags),
		untransitioned: log.New(w, "fix-me: ", log.LstdFlags),
	}
	return
}
func (cl *ConsoleLogger) Item(item interface{}) {
	switch s := item.(type) {
	case fmt.Stringer:
		cl.l.Print(s)
	case string:
		cl.l.Print(s)
	case error:
		cl.l.Print(s)
	default:
		jobj := MakeJObj(item)
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.Encode(jobj)
		cl.l.Print(buf.String())
	}

}
func (cl *ConsoleLogger) Println(items ...interface{}) {
	cl.untransitioned.Println(items...)
}
func (cl *ConsoleLogger) Panic(item interface{}) {
	cl.l.Panic(item)
}

// MongoLogger will archive log items into the specified mongo database.
// It is best of MongoLogger is given a capped collection. However, it
// can do nothing itself to ensure this, with the current state of mgo.
// The collection must be created as capped before any data is inserted.
type MongoLogger struct {
	session         *mgo.Session
	dbName, colName string
	uuid            string

	untransitioned *log.Logger
}

func NewMongoLogger(addr, dbName, collectionName, uuid string) (ml *MongoLogger, err error) {
	ml = &MongoLogger{
		dbName:         dbName,
		colName:        collectionName,
		untransitioned: log.New(os.Stderr, "fix-me: ", log.LstdFlags),
		uuid:           uuid,
	}
	ml.session, err = mgo.Dial(addr)
	if err != nil {
		ml.session = nil
	}
	return
}
func (ml *MongoLogger) Item(item interface{}) {
	if ml.session == nil {
		return
	}
	jobj := MakeJObj(item)
	jobj["uuid"] = ml.uuid

	db := ml.session.DB(ml.dbName)
	col := db.C(ml.colName)

	err := col.Insert(jobj)
	if err != nil {
		log.Printf("Could not log %v: %v", jobj, err)
	}
}
func (ml *MongoLogger) Println(items ...interface{}) {
	ml.Item(fmt.Sprint(items...))
	ml.untransitioned.Println(items...)
}
func (ml *MongoLogger) Panic(item interface{}) {
	if ml.session == nil {
		return
	}
	var strace []string
	for skip := 1; ; skip++ {
		pc, file, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		// if file[len(file)-1] == 'c' {
		// continue
		// }
		f := runtime.FuncForPC(pc)
		strace = append(strace, fmt.Sprintf("%s:%d %s()\n", file, line, f.Name()))
	}
	jobj := map[string]interface{}{
		"Panic": map[string]interface{}{
			"Item":  item,
			"Trace": strace,
		},
		"uuid": ml.uuid,
	}

	db := ml.session.DB(ml.dbName)
	col := db.C(ml.colName)

	err := col.Insert(jobj)
	if err != nil {
		log.Printf("Could not log %v: %v", jobj, err)
	}

	panic(item)
}
