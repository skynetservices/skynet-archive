package skynet

import (
	"fmt"
	"log"
	"os"
)

// FileSemanticLogger logs logging data to files... semantically!
type FileSemanticLogger struct {
	log *log.Logger
}

// NewFileSemanticLogger creates a new logger with the given name that
// logs to the given filename
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
func (fl *FileSemanticLogger) Log(payload *LogPayload) {
	// TODO: Consider using more payload fields
	fl.log.Printf("%v: %s\n", payload.Level, payload.Message)
}

func (fl *FileSemanticLogger) Trace(msg string) {
	// TODO: Consider using more payload fields
	fl.log.Printf("%v: %s\n", TRACE, msg)
}

func (fl *FileSemanticLogger) Debug(msg string) {
	// TODO: Consider using more payload fields
	fl.log.Printf("%v: %s\n", DEBUG, msg)
}

func (fl *FileSemanticLogger) Info(msg string) {
	// TODO: Consider using more payload fields
	fl.log.Printf("%v: %s\n", INFO, msg)
}

func (fl *FileSemanticLogger) Warn(msg string) {
	// TODO: Consider using more payload fields
	fl.log.Printf("%v: %s\n", WARN, msg)
}

func (fl *FileSemanticLogger) Error(msg string) {
	// TODO: Consider using more payload fields
	fl.log.Printf("%v: %s\n", ERROR, msg)
}

// Fatal populates payload.StackTrace then panics
func (fl *FileSemanticLogger) Fatal(msg string) {
	payload := NewLogPayload(FATAL, msg)
	payload.SetException()
	fl.Log(payload)
	panic(payload)
}

// BenchmarkInfo currently does nothing but should measure the time
// it takes to execute `f` based on the log level
func (fl *FileSemanticLogger) BenchmarkInfo(level LogLevel, msg string,
	f func(logger SemanticLogger)) {
	// TODO: Implement
}
