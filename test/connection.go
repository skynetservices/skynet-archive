package test

import (
	"github.com/skynetservices/skynet2"
	"time"
)

type Connection struct {
	SetIdleTimeoutFunc func(timeout time.Duration)
	AddrFunc           func() string

	CloseFunc    func()
	IsClosedFunc func() bool

	SendFunc        func(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) (err error)
	SendTimeoutFunc func(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}, timeout time.Duration) (err error)
}

func (c *Connection) SetIdleTimeout(timeout time.Duration) {
	if c.SetIdleTimeoutFunc != nil {
		c.SetIdleTimeoutFunc(timeout)
	}
}

func (c *Connection) Addr() string {
	if c.AddrFunc != nil {
		return c.AddrFunc()
	}

	return ""
}

func (c *Connection) Close() {
	if c.CloseFunc != nil {
		c.CloseFunc()
	}
}

func (c *Connection) IsClosed() bool {
	if c.IsClosedFunc != nil {
		return c.IsClosedFunc()
	}

	return false
}

func (c *Connection) Send(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) (err error) {
	if c.SendFunc != nil {
		return c.SendFunc(ri, fn, in, out)
	}

	return nil
}

func (c *Connection) SendTimeout(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}, timeout time.Duration) (err error) {
	if c.SendTimeoutFunc != nil {
		return c.SendTimeoutFunc(ri, fn, in, out, timeout)
	}

	return nil
}
