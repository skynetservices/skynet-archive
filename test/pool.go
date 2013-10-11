package test

import (
	"github.com/skynetservices/skynet2"
	"github.com/skynetservices/skynet2/client/conn"
)

type Pool struct {
	AddInstanceFunc    func(s skynet.ServiceInfo)
	UpdateInstanceFunc func(s skynet.ServiceInfo)
	RemoveInstanceFunc func(s skynet.ServiceInfo)

	AcquireFunc func(s skynet.ServiceInfo) (conn.Connection, error)
	ReleaseFunc func(conn.Connection)

	CloseFunc          func()
	NumInstancesFunc   func() int
	NumConnectionsFunc func() int
}

func (p *Pool) AddInstance(s skynet.ServiceInfo) {
	if p.AddInstanceFunc != nil {
		p.AddInstanceFunc(s)
	}
}

func (p *Pool) UpdateInstance(s skynet.ServiceInfo) {
	if p.UpdateInstanceFunc != nil {
		p.UpdateInstanceFunc(s)
	}
}

func (p *Pool) RemoveInstance(s skynet.ServiceInfo) {
	if p.RemoveInstanceFunc != nil {
		p.RemoveInstanceFunc(s)
	}
}

func (p *Pool) Acquire(s skynet.ServiceInfo) (conn.Connection, error) {
	if p.AcquireFunc != nil {
		return p.AcquireFunc(s)
	}

	return nil, nil
}

func (p *Pool) Release(c conn.Connection) {
	if p.ReleaseFunc != nil {
		p.ReleaseFunc(c)
	}
}

func (p *Pool) Close() {
	if p.CloseFunc != nil {
		p.CloseFunc()
	}
}

func (p *Pool) NumInstances() int {
	if p.NumInstancesFunc != nil {
		return p.NumInstancesFunc()
	}

	return 0
}

func (p *Pool) NumConnections() int {
	if p.NumConnectionsFunc != nil {
		return p.NumConnectionsFunc()
	}

	return 0
}
