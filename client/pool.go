package client

import (
	"errors"
	"github.com/skynetservices/skynet2"
	"github.com/skynetservices/skynet2/client/conn"
	"github.com/skynetservices/skynet2/pools"
	"sync"
)

var UnknownService = errors.New("Service not known to connection pool")

type ConnectionPooler interface {
	AddInstance(s skynet.ServiceInfo)
	UpdateInstance(s skynet.ServiceInfo)
	RemoveInstance(s skynet.ServiceInfo)

	Acquire(s skynet.ServiceInfo) (conn.Connection, error)
	Release(conn.Connection)

	Close()
	NumInstances() int
	NumConnections() int
}

/*
client.Pool Manages connection pools to services
*/
type Pool struct {
	servicePools       map[string]*servicePool
	addInstanceChan    chan skynet.ServiceInfo
	updateInstanceChan chan skynet.ServiceInfo
	removeInstanceChan chan skynet.ServiceInfo
	closeChan          chan bool
	closeWait          sync.WaitGroup
}

/*
client.NewPool returns a new connection pool
*/
func NewPool() *Pool {
	p := &Pool{
		servicePools:       make(map[string]*servicePool),
		addInstanceChan:    make(chan skynet.ServiceInfo, 10),
		updateInstanceChan: make(chan skynet.ServiceInfo, 10),
		removeInstanceChan: make(chan skynet.ServiceInfo, 10),
		closeChan:          make(chan bool),
	}

	go p.mux()

	return p
}

type servicePool struct {
	service skynet.ServiceInfo
	pool    *pools.ResourcePool
}

func (sp *servicePool) Close() {
	sp.pool.Close()
}

func (sp *servicePool) NumResources() int {
	return sp.pool.NumResources()
}

func (p *Pool) mux() {
	for {
		select {
		case i := <-p.addInstanceChan:
			p.addInstanceMux(i)
		case i := <-p.removeInstanceChan:
			p.removeInstanceMux(i)
		case i := <-p.updateInstanceChan:
			p.updateInstanceMux(i)
		case <-p.closeChan:
			p.closeMux()
			return
		}
	}
}

/*
Pool.AddInstance adds connections to instance to the pool
*/
func (p *Pool) AddInstance(s skynet.ServiceInfo) {
	go func() {
		p.addInstanceChan <- s
	}()
}

func (p *Pool) addInstanceMux(s skynet.ServiceInfo) {
	if _, ok := p.servicePools[s.AddrString()]; !ok {
		sp := &servicePool{
			service: s,
			pool: pools.NewResourcePool(func() (pools.Resource, error) {
				c, err := conn.NewConnection(s.Name, GetNetwork(), s.AddrString(), DIAL_TIMEOUT)

				if err != nil {
					c.SetIdleTimeout(config.IdleTimeout)
				}

				return c, err
			},
				config.IdleConnectionsToInstance,
				config.MaxConnectionsToInstance),
		}

		p.servicePools[s.AddrString()] = sp
	} else {
		p.UpdateInstance(s)
	}
}

/*
Pool.UpdateInstance updates information about instance, if it's unknown to the pool it will add it
*/
func (p *Pool) UpdateInstance(s skynet.ServiceInfo) {
	go func() {
		p.updateInstanceChan <- s
	}()
}

func (p *Pool) updateInstanceMux(s skynet.ServiceInfo) {
	if _, ok := p.servicePools[s.AddrString()]; !ok {
		p.AddInstance(s)
		return
	}

	p.servicePools[s.AddrString()].service = s
}

/*
Pool.RemoveInstance removes this instance from the pool and closes all it's connections
*/
func (p *Pool) RemoveInstance(s skynet.ServiceInfo) {
	go func() {
		p.removeInstanceChan <- s
	}()
}

func (p *Pool) removeInstanceMux(s skynet.ServiceInfo) {
	delete(p.servicePools, s.AddrString())
}

/*
Pool.Acquire will return an idle connection or a new one
*/
func (p *Pool) Acquire(s skynet.ServiceInfo) (c conn.Connection, err error) {
	if _, ok := p.servicePools[s.AddrString()]; !ok {
		return nil, UnknownService
	}

	r, err := p.servicePools[s.AddrString()].pool.Acquire()

	if err != nil {
		return nil, err
	}

	return r.(conn.Connection), nil
}

/*
Pool.Release will release a resource for use by others. If the idle queue is
full, the resource will be closed.
*/
func (p *Pool) Release(c conn.Connection) {
	if _, ok := p.servicePools[c.Addr()]; !ok {
		c.Close()
		return
	}

	p.servicePools[c.Addr()].pool.Release(c)
}

/*
Pool.Close will close all network connections associated with all known services
*/
func (p *Pool) Close() {
	p.closeWait.Add(1)
	p.closeChan <- true

	p.closeWait.Wait()
}

func (p *Pool) closeMux() {
	for k, sp := range p.servicePools {
		sp.Close()
		delete(p.servicePools, k)
	}

	p.closeWait.Done()
}

/*
Pool.NumConnections will return the total number of connections across all instances
as many connections could be opening and closing this is an estimate
*/
func (p *Pool) NumConnections() (count int) {
	for _, sp := range p.servicePools {
		count += sp.NumResources()
	}

	return count
}

/*
Pool.NumInstances will return the number of unique instances it's maintaining connections too
*/
func (p *Pool) NumInstances() int {
	return len(p.servicePools)
}
