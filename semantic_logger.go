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

// Payload stores detailed logging information, including all fields
// used by Clarity Service's semantic_logger (see
// https://github.com/ClarityServices/semantic_logger). See the
// Payload struct's inline comments for instructions as to which
// fields should be populated by whom or what (e.g., the user
// (manually), helper functions, or methods on the various loggers in
// this package.)
// Valid log levels -- a list of which is stored in the `LogLevels`
// slice -- include TRACE, DEBUG, INFO, WARN, ERROR, and FATAL.
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

// Exception formats the payload just as
// github.com/ClarityServices/semantic_logger formats Exceptions for
// logging. This package has no Exception data type; all relevant data
// should be stored in a *Payload. The payload's "exception" data is
// generated from a panic's stacktrace using the `genStacktrace`
// helper function.
func (payload *Payload) Exception() string {
	// message << " -- " << "#{exception.class}: #{exception.message}\n
	// #{(exception.backtrace || []).join("\n")}"
	formatStr := "%s -- %s: %s\n%s"
	backtrace := strings.Join(payload.backtrace, "\n")
	return fmt.Sprintf(formatStr, payload.Message, "panic",
		payload.Message, backtrace)
}

// setUnexportedPayloadFields sets the `pid`, `time`, and `hostname`
// fields of the given payload. See the documentation on the Payload
// type for which fields should be set where, and by whom (the user)
// or what (a function or method).
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

// LogLevels are ints for the sake of having a well-defined
// ordering. This is useful for viewing logs more or less severe than
// a given log level. See the LogLevel.LessSevereThan method.
type LogLevel int

const (
	TRACE LogLevel = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
)

// LogLevel stores the valid log levels as specified by
// github.com/ClarityServices/semantic_logger.
var LogLevels = []LogLevel{
	TRACE, DEBUG, INFO, WARN, ERROR, FATAL,
}

// LessSevereThan tells you whether or not `level` is a less severe
// LogLevel than `level2`. This is useful for determining which logs
// to view.
func (level LogLevel) LessSevereThan(level2 LogLevel) bool {
	return int(level) < int(level2)
}

// String helps make LogLevel's more readable by representing them as
// strings instead of ints 0 (TRACE) through 5 (FATAL).
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

// NOTE: The data type names are what they are (and rather verbose) in
// part so that "SemanticLogger" can be replaced with "Logger" in this
// file once the contents of logger.go is no longer needed.

// SemanticLogger is meant to match the format and functionality of
// github.com/ClarityServices/semantic_logger
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
// This section defines methods necessary for MultiSemanticLogger to
// implement SemanticLogger
//

// Log calls .Log(payload) for each logger in the
// MultiSemanticLogger. For each logger, logging behavior may vary
// depending upon the LogLevel.
func (ml MultiSemanticLogger) Log(level LogLevel, msg string,
	payload *Payload) error {
	switch level {
	case TRACE, DEBUG, INFO, WARN, ERROR, FATAL:
		for _, lgr := range ml {
			// Note that this won't work very well if
			// payload.LogLevel == FATAL; the first logger will panic.
			if err := lgr.Log(payload); err != nil {
				log.Printf("Error calling .Log: %v\n", err)
			}
		}
	}
	return nil
}

// BenchmarkInfo runs .BenchmarkInfo(level, msg, f) on every logger in
// the MultiSemanticLogger
func (ml MultiSemanticLogger) BenchmarkInfo(level LogLevel, msg string,
	f func(logger SemanticLogger)) {
	for _, lgr := range ml {
		lgr.BenchmarkInfo(level, msg, f)
	}
}

//
// ConsoleSemanticLogger
//

// ConsoleSemanticLogger logs to the console. True story.
type ConsoleSemanticLogger struct {
	log *log.Logger
}

// NewConsoleSemanticLogger returns a *ConsoleSemanticLogger with the
// given name that logs to the given io.Writer (usually os.Stdin or
// os.Stderr).
func NewConsoleSemanticLogger(name string, w io.Writer) *ConsoleSemanticLogger {
	cl := ConsoleSemanticLogger{
		// TODO: Set this format to match Clarity's Ruby SemanticLogger
		log: log.New(w, fmt.Sprintf("%s: ", name), log.LstdFlags),
	}
	return &cl
}

// Log uses select parts of the given payload and logs it to the
// console.
func (cl *ConsoleSemanticLogger) Log(payload *Payload) error {
	// TODO: Consider using more payload fields
	cl.log.Printf("%v: %s\n", payload.Level, payload.Message)
	return nil
}

// Fatal logs the given payload to the console, then panics.
func (cl *ConsoleSemanticLogger) Fatal(payload *Payload) {
	err := cl.Log(payload)
	if err != nil {
		fmt.Printf("Error logging payload to cl\n")
	}
	cl.log.Fatal(payload)
}

// BenchmarkInfo currently does nothing but should measure the time
// it takes to execute `f` based on the log level.
func (cl *ConsoleSemanticLogger) BenchmarkInfo(level LogLevel, msg string,
	f func(logger SemanticLogger)) {
	// TODO: Implement
}


//
// MongoSemanticLogger
//

// MongoSemanticLogger saves logging data to a MongoDB instance.
type MongoSemanticLogger struct {
	session         *mgo.Session
	dbName, colName string
	uuid            string
}

// NewMongoSemanticLogger connects to a MongoDB instance at the given
// address (often "localhost").
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

// Log saves all fields of the given payload to MongoDB, setting
// unexported fields as necessary. May behave differently based upon
// payload.LogLevel.
func (ml *MongoSemanticLogger) Log(payload *Payload) error {
	// Sanity checks
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
	case FATAL: // User should call `ml.Fatal(payload)` directly
		ml.Fatal(payload)
	default: // Payloads with custom log levels should be logged
		fallthrough
	case TRACE, DEBUG, INFO, WARN, ERROR:
		if payload != nil {
			err := col.Insert(payload)
			if err != nil {
				errStr := "Error logging with MongoSemanticLogger %s: %v"
				return fmt.Errorf(errStr, ml.uuid, err)
			}
		}
	}
	return nil
}

// Fatal logs the given payload to MongoDB, then panics.
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

// BenchmarkInfo currently does nothing but should measure the time
// it takes to execute `f` based on the log level
func (ml *MongoSemanticLogger) BenchmarkInfo(level LogLevel, msg string,
	f func(logger SemanticLogger)) {
	// TODO: Implement
}


// genStacktrace is a helper function for generating stacktrace
// data. Used to populate payload.backtrace
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

//
// FileSemanticLogger
//

// FileSemanticLogger logs logging data to files... semantically!
type FileSemanticLogger struct {
	log *log.Logger
}

// NewMongoSemanticLogger creates a new logger with the given name
// that logs to the given filename
func NewFileSemanticLogger(name, filename string) (*FileSemanticLogger, error) {
	// Open file with append permissions
	flags := os.O_APPEND | os.O_WRONLY | os.O_CREATE
	file, err := os.OpenFile(filename, flags, 0666)
	if err != nil {
		return nil, fmt.Errorf("Error opening '%v': %v", filename, err)
	}
	// Oddity: `file.Close()` never gets called. Everything seems to work.
	fl := FileSemanticLogger{
		log: log.New(file, fmt.Sprintf("%s: ", name), log.LstdFlags),
	}
	return &fl, nil
}

// Log uses select parts of the given payload and logs to fl.log
func (fl *FileSemanticLogger) Log(payload *Payload) error {
	// TODO: Consider using more payload fields
	fl.log.Printf("%v: %s\n", payload.Level, payload.Message)
	return nil
}

// Fatal populates payload.backtrace then panics
func (fl *FileSemanticLogger) Fatal(payload *Payload) {
	payload.backtrace = genStacktrace()
	// Should this call `fl.Log(payload)` before panicking?
	panic(payload)
}

// BenchmarkInfo currently does nothing but should measure the time
// it takes to execute `f` based on the log level
func (fl *FileSemanticLogger) BenchmarkInfo(level LogLevel, msg string,
	f func(logger SemanticLogger)) {
	// TODO: Implement
}
