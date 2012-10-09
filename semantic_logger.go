package skynet

import (
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

type SemanticLogger interface {
	Trace(msg string, payload *Payload, exception ...*Exception)
	Debug(msg string, payload *Payload, exception ...*Exception)
	Info(msg string, payload *Payload, exception ...*Exception)
	Warn(msg string, payload *Payload, exception ...*Exception)
	Error(msg string, payload *Payload, exception ...*Exception)
	Fatal(msg string, payload *Payload, exception ...*Exception)
	BenchmarkInfo(msg, f func(logger SemanticLogger))
}
