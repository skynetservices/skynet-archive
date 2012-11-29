type ServiceClientStub struct {
	SetTimeoutFunc func(retry, giveup time.Duration)
	GetTimeoutFunc func() (retry, giveup time.Duration)
	SendFunc       func(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) (err error)
	SendOnceFunc   func(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) (err error)
}

func (sc *ServiceClientStub) SetTimeout(retry, giveup time.Duration) {
	if sc.SetTimeoutFunc != nil {
		sc.SetTimeoutFunc(retry, giveup)
	}

	return
}

func (sc *ServiceClientStub) GetTimeout() (retry, giveup time.Duration) {
	if sc.GetTimeoutFunc != nil {
		return sc.GetTimeoutFunc()
	}

	return
}

func (sc *ServiceClientStub) Send(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) (err error) {
	// Track that this method was called
	if sc.SendFunc != nil {
		return sc.SendFunc(ri, fn, in, out)
	}

	return
}

func (sc *ServiceClientStub) SendOnce(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) (err error) {
	if sc.SendOnceFunc != nil {
		return sc.SendOnceFunc(ri, fn, in, out)
	}

	return
}

