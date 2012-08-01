package pools

type Resource interface {
	Close()
	IsClosed() bool
}

type Factory func() (Resource, error)

type ResourcePool struct {
	factory       Factory
	idleResources chan Resource
}

func NewResourcePool(factory Factory, idleCapacity int) (rp *ResourcePool) {
	rp = &ResourcePool{
		factory:       factory,
		idleResources: make(chan Resource, idleCapacity),
	}

	return
}

func (rp *ResourcePool) Close() {
	close(rp.idleResources)
}

// ClaimPool() will claim all idle resources from the other pool.
func (rp *ResourcePool) ClaimPool(o *ResourcePool) {
	go func(o *ResourcePool) {
		for resource := range o.idleResources {
			rp.Release(resource)
		}
	}(o)
}

// AcquireOrCreate() will get one of the idle resources, or create a new one.
func (rp *ResourcePool) AcquireOrCreate() (resource Resource, err error) {
	select {
	case resource = <-rp.idleResources:
	default:
		resource, err = rp.factory()
	}
	return
}

// Release() will release a resource for use by others. If the idle queue is
// full, the resource will be closed.
func (rp *ResourcePool) Release(resource Resource) {
	if resource.IsClosed() {
		// don't put it back in the pool.
		return
	}
	select {
	case rp.idleResources <- resource:
	default:
		resource.Close()
	}
}
