package skynet

import (
	"fmt"
	"io"
	"labix.org/v2/mgo"
	"log"
	"runtime"
	"strings"
	"time"
)

type Payload struct {
	Name        string        `json:"name"`
	Application string        `json:"application"`
	HostName    string        `json:"host_name"`
	ThreadName  string        `json:"thread_name"`
	Message     string        `json:"message"`
	Tags        []string      `json:"tags"`
	PID         int           `json:"pid"`
	Level       LogLevel      `json:"level"`
	Time        time.Time     `json:"time"`
	Duration    time.Duration `json:"duration"`
	Table       string        `json:"table"`
	Action      string        `json:"action"`
	UUID        string        `json:"uuid"`
	Backtrace   []string      `json:"backtrace"`
}

func (pl *Payload) Exception() string {
	// message << " -- " << "#{exception.class}: #{exception.message}\n
	// #{(exception.backtrace || []).join("\n")}"
	formatStr := "%s -- %s: %s\n%s"
	backtrace := strings.Join(pl.Backtrace, "\n")
	return fmt.Sprintf(formatStr, pl.Message, "panic", pl.Message, backtrace)
}

type LogLevel int

const (
	TRACE LogLevel = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
)

var LogLevels = []LogLevel{
	TRACE, DEBUG, INFO, WARN, ERROR, FATAL,
}

func (level LogLevel) LessSevereThan(level2 LogLevel) bool {
	return int(level) < int(level2)
}

func (level LogLevel) String() string {
	switch level {
	case TRACE:
		return "TRACE"
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	}
	return "CUSTOM"
}

// Goal: When done replicating `Logger` logic as `SemanticLogger`s,
// s/SemanticLogger/Logger/ in this file

type SemanticLogger interface {
	Log(payload *Payload) error
	Fatal(payload *Payload)
	BenchmarkInfo(level LogLevel, msg string, f func(logger SemanticLogger))
}

type MultiSemanticLogger []SemanticLogger

func NewMultiSemanticLogger(loggers ...SemanticLogger) (ml MultiSemanticLogger) {
	ml = loggers
	return
}

//
// Define methods necessary for MultiSemanticLogger to implement
// SemanticLogger
//

func (ml MultiSemanticLogger) Log(level LogLevel, msg string,
	payload *Payload) error {
	switch level {
	case TRACE, DEBUG, INFO, WARN, ERROR, FATAL:
		for _, lgr := range ml {
			if err := lgr.Log(payload); err != nil {
				log.Printf("Error calling .Log: %v\n", err)
			}
		}
	}
	return nil
}

func (ml MultiSemanticLogger) BenchmarkInfo(level LogLevel, msg string,
	f func(logger SemanticLogger)) {
	for _, lgr := range ml {
		lgr.BenchmarkInfo(level, msg, f)
	}
}

//
// ConsoleSemanticLogger
//

type ConsoleSemanticLogger struct {
	l *log.Logger
}

func NewConsoleSemanticLogger(name string, w io.Writer) *ConsoleSemanticLogger {
	cl := ConsoleSemanticLogger{
		// TODO: Set this format to match Clarity's Ruby SemanticLogger
		l: log.New(w, fmt.Sprintf("%s: ", name), log.LstdFlags),
	}
	return &cl
}

func (cl *ConsoleSemanticLogger) Log(payload *Payload) error {
	cl.l.Printf("%v: %s\n", payload.Level, payload.Message)
	return nil
}

func (cl *ConsoleSemanticLogger) Fatal(payload *Payload) {
	cl.l.Fatal(payload)
}

func (cl *ConsoleSemanticLogger) BenchmarkInfo(level LogLevel, msg string, f func(logger SemanticLogger)) {
	// TODO: Implement
}


//
// MongoSemanticLogger
//

type MongoSemanticLogger struct {
	session         *mgo.Session
	dbName, colName string
	uuid            string
}

func NewMongoSemanticLogger(addr, dbName, collectionName,
	uuid string) (ml *MongoSemanticLogger, err error) {
	ml = &MongoSemanticLogger{
		dbName:         dbName,
		colName:        collectionName,
		uuid:           uuid,
	}
	ml.session, err = mgo.Dial(addr)
	return
}

func (ml *MongoSemanticLogger) Log(payload *Payload) error {
	if ml == nil {
		return fmt.Errorf("Can't log to nil *MongoSemanticLogger")
	}
	if payload == nil {
		return fmt.Errorf("Can't log nil *Payload")
	}

	db := ml.session.DB(ml.dbName)
	col := db.C(ml.colName)

	// TODO: Remove once basics are in place (MongoDB logging, etc);
	// `switch` for testing purposes only
	switch payload.Level {
	case TRACE, DEBUG, INFO, WARN, ERROR: // Use Payload
		if payload != nil {
			err := col.Insert(payload)
			if err != nil {
				errStr := "Error logging with MongoSemanticLogger %s: %v"
				return fmt.Errorf(errStr, ml.uuid, err)
			}
		}
	case FATAL: // Should call ml.Fatal(payload) directly
		ml.Fatal(payload)
	}
	return nil
}

func (ml *MongoSemanticLogger) Fatal(payload *Payload) {
	db := ml.session.DB(ml.dbName)
	col := db.C(ml.colName)

	var stackTrace []string

	for skip := 1; ; skip++ {
		pc, file, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		f := runtime.FuncForPC(pc)
		traceLine := fmt.Sprintf("%s:%d %s()\n", file, line, f.Name())
		stackTrace = append(stackTrace, traceLine)
	}

	payload.Backtrace = stackTrace
	err := col.Insert(payload)
	if err != nil {
		log.Printf("Logging error: %v", err)
	}
	panic(payload)
}

func (ml *MongoSemanticLogger) BenchmarkInfo(level LogLevel, msg string,
		f func(logger SemanticLogger)) {
		// TODO: Implement
}
