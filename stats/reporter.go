package stats

import (
	"time"
)

var reporters []Reporter

type Reporter interface {
	UpdateHostStats(host string, stats Host)
	MethodCalled(method string)
	MethodCompleted(method string, duration time.Duration, err error)
}

func AddReporter(r Reporter) {
	reporters = append(reporters, r)
}

func UpdateHostStats(host string, s Host) {
	for _, r := range reporters {
		go r.UpdateHostStats(host, s)
	}
}

func MethodCalled(method string) {
	for _, r := range reporters {
		go r.MethodCalled(method)
	}
}

func MethodCompleted(method string, duration time.Duration, err error) {
	for _, r := range reporters {
		go r.MethodCompleted(method, duration, err)
	}
}
