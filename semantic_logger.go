package skynet

import (
	"fmt"
	"log"
	"os"
	"runtime"
	// "strings"
	"time"
)

// LogPayload stores detailed logging information, including all
// fields used by Clarity Service's semantic_logger (see
// https://github.com/ClarityServices/semantic_logger). See the
// LogPayload struct's inline comments for instructions as to which
// fields should be populated by whom or what (e.g., the user
// (manually), helper functions, or methods on the various loggers in
// this package.)
// Valid log levels -- a list of which is stored in the `LogLevels`
// slice -- include TRACE, DEBUG, INFO, WARN, ERROR, and FATAL.
type LogPayload struct {
	// Set by user by passing values to NewLogPayload()
	Level   LogLevel `json:"level" bson:"level"`
	Message string   `json:"message" bson:"message"`
	// Set automatically within NewLogPayload()
	LevelIndex int    `json:"level_index" bson:"level_index"`
	Name       string `json:"name" bson:"name"` // Name of calling function
	// Set by .setKnownFields()
	PID      int       `json:"pid" bson:"pid"`
	Time     time.Time `json:"time" bson:"time"`
	HostName string    `json:"host_name" bson:"host_name"`
	// Set by .SetTags() convenience method
	Tags []string `json:"tags" bson:"tags"`
	// Should be set by .Log()
	UUID string `json:"uuid" bson:"uuid"` // Logger's UUID
	// Should be set by BenchmarkInfo() if called
	Duration time.Duration `json:"duration" bson:"duration"`
	// Optionally set by user manually
	ThreadName  string      `json:"thread_name" bson:"thread_name"`
	Application string      `json:"application" bson:"application"`
	Payload     interface{} `json:"payload" bson:"payload"` // Arbitrary data
	Exception   *Exception  `json:"exception" bson:"exception"`
}

type Exception struct {
	Name    string `json:"name" bson:"name"`
	Message string `json:"message" bson:"message"`
	// Set by Fatal() method if need be
	StackTrace []string `json:"stack_trace" bson:"stack_trace"`
}

// SetException sets the given payload's `Exception` field. The
// payload's "exception" data is generated from a panic's stacktrace
// using the `genStacktrace` helper function.
func (payload *LogPayload) SetException() {
	// TODO: If the following logging format is still used in by
	// github.com/ClarityServices/semantic_logger, use it here:
	// message << " -- " << "#{exception.class}: #{exception.message}\n
	// #{(exception.backtrace || []).join("\n")}"

	// ...then use this code to fill .Message with the above-formatted
	// logging information:
	// formatStr := "%s -- %s: %s\n%s"
	// stacktrace := strings.Join(payload.Exception.StackTrace, "\n")
	// payload.Exception.Message = fmt.Sprintf(formatStr, payload.Message, "panic",
	// 	payload.Message, stacktrace)

	payload.Exception = &Exception{
		// TODO: Decide what `Name` should be. Go doesn't have
		// exceptions, and therefore has no exceptions whose names we
		// can put here.
		Name:       "",
		Message:    payload.Message,
		StackTrace: genStacktrace(),
	}
}

// setKnownFields sets the `Application`, `PID`, `Time`, and
// `HostName` fields of the given payload. See the documentation on
// the LogPayload type for which fields should be set where, and by
// whom (the user) or what (a function or method).
func (payload *LogPayload) setKnownFields() {
	payload.PID = os.Getpid()
	payload.Time = time.Now()
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("Error getting hostname: %v\n", err)
	}
	payload.HostName = hostname
}

// SetTags is a convenience method for adding tags to *LogPayload's,
// since `payload.SetTags("tag1", "tag2")` is cleaner than
// `payload.Tags = []string{"tag1", "tag2"}`
func (payload *LogPayload) SetTags(tags ...string) {
	payload.Tags = tags
}

// NewLogPayload is a convenience function for creating *LogPayload's
func NewLogPayload(level LogLevel, formatStr string, vars ...interface{}) *LogPayload {
	payload := &LogPayload{
		Level:      level,
		LevelIndex: levelIndex(level),
		Message:    fmt.Sprintf(formatStr, vars...),
		// 1 == skynet.NewLogPayload
		// 2 == skynet.(*MongoSemanticLogger).Fatal
		// 3 == What we want
		// 4 (or shortly thereafter) == main.main
		Name: getCallerName(3),
	}
	// payload.setKnownFields() called in .Log() method; not calling here

	// TODO: Come up with a way to intelligently auto-fill ThreadName,
	// if possible

	return payload
}

func getCallerName(skip int) string {
	pc, _, _, _ := runtime.Caller(skip)
	f := runtime.FuncForPC(pc)
	return f.Name()
}

type LogLevel string

const (
	TRACE LogLevel = "trace"
	DEBUG LogLevel = "debug"
	INFO  LogLevel = "info"
	WARN  LogLevel = "warn"
	ERROR LogLevel = "error"
	FATAL LogLevel = "fatal"
)

// LogLevel stores the valid log levels as specified by
// github.com/ClarityServices/semantic_logger. Its index corresponds
// to the log level it represents. (LogLevels[0] == "trace", ...,
// LogLevels[5] == "fatal")
var LogLevels = []LogLevel{
	TRACE, DEBUG, INFO, WARN, ERROR, FATAL,
}

// levelIndex turns the given level into its corresponding integer
// value. "trace" == 0, "debug" == 1, ... "fatal" == 5
func levelIndex(level LogLevel) int {
	switch level {
	case TRACE:
		return 0
	case DEBUG:
		return 1
	case INFO:
		return 2
	case WARN:
		return 3
	case ERROR:
		return 4
	case FATAL:
		return 5
	}
	return -1
}

// LessSevereThan tells you whether or not `level` is a less severe
// LogLevel than `level2`. This is useful for determining which logs
// to view.
func (level LogLevel) LessSevereThan(level2 LogLevel) bool {
	return levelIndex(level) < levelIndex(level2)
}

// NOTE: The data type names are what they are (and rather verbose) in
// part so that "SemanticLogger" can be replaced with "Logger" in this
// file once the contents of logger.go is no longer needed.

// SemanticLogger is meant to match the format and functionality of
// github.com/ClarityServices/semantic_logger
type SemanticLogger interface {
	Log(payload *LogPayload)
	Trace(msg string)
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Fatal(msg string)
	BenchmarkInfo(level LogLevel, msg string, f func(logger SemanticLogger))
}

// genStacktrace is a helper function for generating stacktrace
// data. Used to populate (*LogPayload).StackTrace
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
