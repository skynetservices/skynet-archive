package sleeper

import (
	"time"
)

type Request struct {
	Duration           time.Duration
	UnregisterWhenDone bool
	ExitWhenDone       bool
	PanicWhenDone      bool
	Message            string
}

type Response struct {
	Message string
}
