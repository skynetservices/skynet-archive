package skynet

import (
	"fmt"
	"io"
	"log"
)

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
func (cl *ConsoleSemanticLogger) Log(payload *LogPayload) {
	// TODO: Consider using more payload fields
	cl.log.Printf("%v: %s\n", payload.Level, payload.Message)
}

func (cl *ConsoleSemanticLogger) Trace(msg string) {
	// TODO: Consider using more payload fields
	cl.log.Printf("%v: %s\n", TRACE, msg)
}

func (cl *ConsoleSemanticLogger) Debug(msg string) {
	// TODO: Consider using more payload fields
	cl.log.Printf("%v: %s\n", DEBUG, msg)
}

func (cl *ConsoleSemanticLogger) Info(msg string) {
	// TODO: Consider using more payload fields
	cl.log.Printf("%v: %s\n", INFO, msg)
}

func (cl *ConsoleSemanticLogger) Warn(msg string) {
	// TODO: Consider using more payload fields
	cl.log.Printf("%v: %s\n", WARN, msg)
}

func (cl *ConsoleSemanticLogger) Error(msg string) {
	// TODO: Consider using more payload fields
	cl.log.Printf("%v: %s\n", ERROR, msg)
}

// Fatal logs the given payload to the console, then panics.
func (cl *ConsoleSemanticLogger) Fatal(msg string) {
	payload := NewLogPayload(FATAL, msg)
	payload.SetException()
	cl.Log(payload)
	panic(payload)
}

// BenchmarkInfo currently does nothing but should measure the time
// it takes to execute `f` based on the log level.
func (cl *ConsoleSemanticLogger) BenchmarkInfo(level LogLevel, msg string,
	f func(logger SemanticLogger)) {
	// TODO: Implement
}
