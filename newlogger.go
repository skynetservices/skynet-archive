package skynet

import (
	"time"
)

type LogPayload struct {
	Duration time.Duration `json:"duration"`
	Result   interface{}   `json:"result"` // What type should this be?
	Table    string        `json:"table"`
	Action   string        `json:"action"`
}

// Will rename to just 'Logger' once transition is complete
type MyLogger interface {
	Trace(msg string, payload ...*LogPayload)
	Debug(msg string, payload ...*LogPayload)
	Info(msg string, payload ...*LogPayload)
	Warn(msg string, payload ...*LogPayload)
	Error(msg string, payload ...*LogPayload)
	Fatal(msg string, payload ...*LogPayload)
}
