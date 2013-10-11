package test

import (
	"github.com/skynetservices/skynet2"
	"time"
)

type ServiceClient struct {
	SetDefaultTimeoutFunc func(retry, giveup time.Duration)
	GetDefaultTimeoutFunc func() (retry, giveup time.Duration)

	CloseFunc func()

	SendFunc     func(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) (err error)
	SendOnceFunc func(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) (err error)

	NotifyFunc  func(n skynet.InstanceNotification)
	MatchesFunc func(n skynet.ServiceInfo) bool
}

func (sc *ServiceClient) SetDefaultTimeout(retry, giveup time.Duration) {
	if sc.SetDefaultTimeoutFunc != nil {
		sc.SetDefaultTimeoutFunc(retry, giveup)
	}

	return
}

func (sc *ServiceClient) GetDefaultTimeout() (retry, giveup time.Duration) {
	if sc.GetDefaultTimeoutFunc != nil {
		return sc.GetDefaultTimeoutFunc()
	}

	return
}

func (sc *ServiceClient) Send(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) (err error) {
	if sc.SendFunc != nil {
		return sc.SendFunc(ri, fn, in, out)
	}

	return
}

func (sc *ServiceClient) SendOnce(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) (err error) {
	if sc.SendOnceFunc != nil {
		return sc.SendOnceFunc(ri, fn, in, out)
	}

	return
}

func (sc *ServiceClient) Close() {
	if sc.CloseFunc != nil {
		sc.CloseFunc()
	}

	return
}

func (sc *ServiceClient) Notify(n skynet.InstanceNotification) {
	if sc.NotifyFunc != nil {
		sc.NotifyFunc(n)
	}

	return
}

func (sc *ServiceClient) Matches(s skynet.ServiceInfo) bool {
	if sc.MatchesFunc != nil {
		return sc.MatchesFunc(s)
	}

	return false
}
