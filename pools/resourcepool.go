package pools

import (
	"errors"
)

type Resource interface {
	Close()
	IsClosed() bool
}

type Factory func() (Resource, error)

type ResourcePool struct {
	factory       Factory
	idleResources ring
	idleCapacity  int
	maxResources  int
	numResources  int

	acqchan chan acquireMessage
	rchan   chan releaseMessage
	cchan   chan closeMessage

	activeWaits []acquireMessage
}

func NewResourcePool(factory Factory, idleCapacity, maxResources int) (rp *ResourcePool) {
	rp = &ResourcePool{
		factory:      factory,
		idleCapacity: idleCapacity,
		maxResources: maxResources,

		acqchan: make(chan acquireMessage),
		rchan:   make(chan releaseMessage, 1),
		cchan:   make(chan closeMessage, 1),
	}

	go rp.mux()

	return
}

type releaseMessage struct {
	r Resource
}

type acquireMessage struct {
	rch chan Resource
	ech chan error
}

type closeMessage struct {
}

func (rp *ResourcePool) mux() {
loop:
	for {
		select {
		case acq := <-rp.acqchan:
			rp.acquire(acq)
		case rel := <-rp.rchan:
			if len(rp.activeWaits) != 0 {
				// someone is waiting - give them the resource if we can
				if !rel.r.IsClosed() {
					rp.activeWaits[0].rch <- rel.r
				} else {
					// if we can't, discard the released resource and create a new one
					r, err := rp.factory()
					if err != nil {
						// reflect the smaller number of existant resources
						rp.numResources--
						rp.activeWaits[0].ech <- err
					} else {
						rp.activeWaits[0].rch <- r
					}
				}
			} else {
				// if no one is waiting, release it for idling or closing
				rp.release(rel.r)
			}

		case _ = <-rp.cchan:
			break loop
		}
	}
	for !rp.idleResources.Empty() {
		rp.idleResources.Dequeue().Close()
	}
	for _, aw := range rp.activeWaits {
		aw.ech <- errors.New("Resource pool closed")
	}
}

func (rp *ResourcePool) acquire(acq acquireMessage) {
	for !rp.idleResources.Empty() {
		r := rp.idleResources.Dequeue()
		if !r.IsClosed() {
			acq.rch <- r
			return
		}
		// discard closed resources
		rp.numResources--
	}
	if rp.maxResources > 0 && rp.numResources >= rp.maxResources {
		// we need to wait until something comes back in
		rp.activeWaits = append(rp.activeWaits, acq)
		return
	}

	r, err := rp.factory()
	if err != nil {
		acq.ech <- err
	} else {
		rp.numResources++
		acq.rch <- r
	}

	return
}

func (rp *ResourcePool) release(resource Resource) {
	if resource.IsClosed() {
		// don't put it back in the pool.
		rp.numResources--
		return
	}
	if rp.idleCapacity != 0 && rp.idleResources.Size() == rp.idleCapacity {
		resource.Close()
		rp.numResources--
		return
	}

	rp.idleResources.Enqueue(resource)
}

// Acquire() will get one of the idle resources, or create a new one.
func (rp *ResourcePool) Acquire() (resource Resource, err error) {
	acq := acquireMessage{
		rch: make(chan Resource),
		ech: make(chan error),
	}
	rp.acqchan <- acq

	select {
	case resource = <-acq.rch:
	case err = <-acq.ech:
	}

	return
}

// Release() will release a resource for use by others. If the idle queue is
// full, the resource will be closed.
func (rp *ResourcePool) Release(resource Resource) {
	rel := releaseMessage{
		r: resource,
	}
	rp.rchan <- rel
}

// Close() closes all the pools resources.
func (rp *ResourcePool) Close() {
	rp.cchan <- closeMessage{}
}
