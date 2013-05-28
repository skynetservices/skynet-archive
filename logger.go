package skynet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
