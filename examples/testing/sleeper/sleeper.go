package sleeper

import (
	"time"
)

type Request struct {
	Duration time.Duration
	Message  string
}

type Response struct {
	Message string
}
