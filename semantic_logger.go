package skynet

import (
	"fmt"
	"io"
	"labix.org/v2/mgo"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

type Payload struct {
	// Set by user
	ThreadName  string        `json:"thread_name"`
	Level       LogLevel      `json:"level"`
	Message     string        `json:"message"`
	Tags        []string      `json:"tags"`
	Action      string        `json:"action"`
	// Set by setUnexportedPayloadFields()
	hostname    string        `json:"host_name"`
	pid         int           `json:"pid"`
	time        time.Time     `json:"time"`
	// Should be set by Log() method
	name        string        `json:"name"`
	uuid        string        `json:"uuid"`
	table       string        `json:"table"` // Set automatically???
	// Set by Fatal() method if need be
	backtrace   []string      `json:"backtrace"`
	// Should be set by BenchmarkInfo() if called
	duration    time.Duration `json:"duration"`
	// TODO: When should payload.Application be set?
	Application string        `json:"application"`
}

func (pl *Payload) Exception() string {
	// message << " -- " << "#{exception.class}: #{exception.message}\n
	// #{(exception.backtrace || []).join("\n")}"
	formatStr := "%s -- %s: %s\n%s"
	backtrace := strings.Join(pl.backtrace, "\n")
	return fmt.Sprintf(formatStr, pl.Message, "panic", pl.Message, backtrace)
}

func setUnexportedPayloadFields(payload *Payload) error {
	payload.pid = os.Getpid()
	payload.time = time.Now()
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("Error getting hostname: %v", err)
	}
	payload.hostname = hostname
	return nil
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
	log *log.Logger
}

func NewConsoleSemanticLogger(name string, w io.Writer) *ConsoleSemanticLogger {
	cl := ConsoleSemanticLogger{
		// TODO: Set this format to match Clarity's Ruby SemanticLogger
		log: log.New(w, fmt.Sprintf("%s: ", name), log.LstdFlags),
	}
	return &cl
}

func (cl *ConsoleSemanticLogger) Log(payload *Payload) error {
	cl.log.Printf("%v: %s\n", payload.Level, payload.Message)
	return nil
}

func (cl *ConsoleSemanticLogger) Fatal(payload *Payload) {
	cl.log.Fatal(payload)
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

	err := setUnexportedPayloadFields(payload)
	if err != nil {
		// Don't return this error; too minor an issue to be worth
		// it. (Onward!)
		errStr := "From setUnexportedPayloadFields for payload '%+v': %v\n"
		log.Printf(errStr, payload, err)
	}
	payload.uuid = ml.uuid
	payload.table = ml.colName

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
	// Log to proper DB and collection name
	db := ml.session.DB(ml.dbName)
	col := db.C(ml.colName)

	// Generate stacktrace, then log to MongoDB before panicking
	payload.backtrace = genStacktrace()
	err := col.Insert(payload)
	if err != nil {
		log.Printf("Error inserting '%+v' into %s collection: %v",
			payload, ml.colName, err)
	}
	panic(payload)
}

func (ml *MongoSemanticLogger) BenchmarkInfo(level LogLevel, msg string,
		f func(logger SemanticLogger)) {
		// TODO: Implement
}

func genStacktrace() (stacktrace []string) {
	// TODO: Make sure that `skip` should begin at 1, not 2
	for skip := 1; ; skip++ {
		pc, file, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		f := runtime.FuncForPC(pc)
		traceLine := fmt.Sprintf("%s:%d %s()\n", file, line, f.Name())
		stacktrace = append(stacktrace, traceLine)
	}
	return
}
