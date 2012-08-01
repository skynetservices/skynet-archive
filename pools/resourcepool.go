package pools

type Resource interface {
	Close()
	IsClosed() bool
}

type Factory func() (Resource, error)

type ResourcePool struct {
	factory       Factory
	idleResources ring
	idleCapacity  int
	maxCreate     int

	acqchan chan acquireMessage
	rchan   chan releaseMessage
	cchan   chan closeMessage
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

func NewResourcePool(factory Factory, idleCapacity, maxCreate int) (rp *ResourcePool) {
	rp = &ResourcePool{
		factory:      factory,
		idleCapacity: idleCapacity,
		maxCreate:    maxCreate,

		acqchan: make(chan acquireMessage, 1),
		rchan:   make(chan releaseMessage, 1),
		cchan:   make(chan closeMessage, 1),
	}

	go rp.mux()

	return
}

func (rp *ResourcePool) mux() {
loop:
	for {
		select {
		case acq := <-rp.acqchan:
			r, err := rp.acquire()
			if err == nil {
				acq.rch <- r
			} else {
				acq.ech <- err
			}
		case rel := <-rp.rchan:
			rp.release(rel.r)
		case _ = <-rp.cchan:
			break loop
		}
	}
	for !rp.idleResources.Empty() {
		rp.idleResources.Dequeue().Close()
	}
}

func (rp *ResourcePool) acquire() (resource Resource, err error) {
	if !rp.idleResources.Empty() {
		resource = rp.idleResources.Dequeue()
		return
	}
	resource, err = rp.factory()
	return
}

func (rp *ResourcePool) release(resource Resource) {
	if resource.IsClosed() {
		// don't put it back in the pool.
		return
	}
	if rp.idleResources.Size() == rp.idleCapacity {
		resource.Close()
		return
	}

	rp.idleResources.Enqueue(resource)
}

// Acquire() will get one of the idle resources, or create a new one.
func (rp *ResourcePool) Acquire() (resource Resource, err error) {
	acq := acquireMessage{
		rch: make(chan Resource, 1),
		ech: make(chan error, 1),
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
