package skynet

import (
	"fmt"
	"io"
	"log"
	"time"
)

type Exception struct {
	Exception string `json:"exception"`
	Message   string `json:"message"`
	Backtrace string `json:"backtrace"`
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
}

// Goal: When done replicating `Logger` logic as `SemanticLogger`s,
// s/SemanticLogger/Logger/ in this file

type SemanticLogger interface {
	Trace(msg string, payload *Payload, exception ...*Exception)
	Debug(msg string, payload *Payload, exception ...*Exception)
	Info(msg string, payload *Payload, exception ...*Exception)
	Warn(msg string, payload *Payload, exception ...*Exception)
	Error(msg string, payload *Payload, exception ...*Exception)
	Fatal(msg string, payload *Payload, exception ...*Exception)
	BenchmarkInfo(msg string, f func(logger SemanticLogger))
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

func (ml MultiSemanticLogger) Trace(msg string, payload *Payload,
	exception ...*Exception) {
	for _, lgr := range ml {
		lgr.Trace(msg, payload, exception...)
	}
}

func (ml MultiSemanticLogger) Debug(msg string, payload *Payload,
	exception ...*Exception) {
	for _, lgr := range ml {
		lgr.Debug(msg, payload, exception...)
	}
}

func (ml MultiSemanticLogger) Info(msg string, payload *Payload,
	exception ...*Exception) {
	for _, lgr := range ml {
		lgr.Info(msg, payload, exception...)
	}
}

func (ml MultiSemanticLogger) Warn(msg string, payload *Payload,
	exception ...*Exception) {
	for _, lgr := range ml {
		lgr.Warn(msg, payload, exception...)
	}
}

func (ml MultiSemanticLogger) Error(msg string, payload *Payload,
	exception ...*Exception) {
	for _, lgr := range ml {
		lgr.Error(msg, payload, exception...)
	}
}

func (ml MultiSemanticLogger) Fatal(msg string, payload *Payload,
	exception ...*Exception) {
	for _, lgr := range ml {
		lgr.Fatal(msg, payload, exception...)
	}
}

func (ml MultiSemanticLogger) BenchmarkInfo(msg, f func(logger SemanticLogger)) {
	for _, lgr := range ml {
		lgr.BenchmarkInfo(msg, f)
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
