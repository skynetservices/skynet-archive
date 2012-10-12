package skynet

import (
	"fmt"
	"io"
	"labix.org/v2/mgo"
	"log"
	"strings"
	"time"
)

type Exception struct {
	Message   string   `json:"message"`
	Backtrace []string `json:"backtrace"`
}

func (excep *Exception) String() string {
	// message << " -- " << "#{exception.class}: #{exception.message}\n
	// #{(exception.backtrace || []).join("\n")}"
	formatStr := "%s -- %s: %s\n%s"
	backtrace := strings.Join(excep.Backtrace, "\n")
	return fmt.Sprintf(formatStr, excep.Message, "panic", excep.Message,
		backtrace)
}

type Payload struct {
	Name        string        `json:"name"`
	Application string        `json:"application"`
	HostName    string        `json:"host_name"`
	ThreadName  string        `json:"thread_name"`
	Message     string        `json:"message"`
	Tags        []string      `json:"tags"`
	PID         int64         `json:"pid"`
	Level       string        `json:"level"`
	Time        time.Time     `json:"time"`
	Duration    time.Duration `json:"duration"`
	Table       string        `json:"table"`
	Action      string        `json:"action"`
	UUID        string        `json:"uuid"`
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
	Log(LogLevel, string, *Payload, *Exception) error
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

func (ml MultiSemanticLogger) Log(level LogLevel, msg string, payload *Payload,
	excep *Exception) error {
	switch level {
	case TRACE, DEBUG, INFO, WARN, ERROR, FATAL:
		for _, lgr := range ml {
			// TODO: Decide what to do with returned `error` value
			lgr.Log(level, msg, payload, excep)
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

func (ml *MongoSemanticLogger) Log(level LogLevel, msg string,
	payload *Payload, excep *Exception) error {
	if ml == nil {
		return fmt.Errorf("Can't log to nil *MongoSemanticLogger")
	}
	db := ml.session.DB(ml.dbName)
	col := db.C(ml.colName)

	// TODO: Remove once basics are in place (MongoDB logging, etc);
	// `switch` for testing purposes only
	switch level {
	case TRACE, DEBUG, INFO, WARN, ERROR: // Use Payload
		if payload != nil {
			err := col.Insert(payload)
			if err != nil {
				errStr := "Error logging with MongoSemanticLogger %s: %v"
				return fmt.Errorf(errStr, ml.uuid, err)
			}
		}
	case FATAL: // Use Exception
	}
	return nil
}
