package roundrobin

import (
	"container/list"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/client/loadbalancer"
	"sync"
)

type LoadBalancer struct {
	instances     map[string]*list.Element
	instanceMutex sync.Mutex
	instanceList  list.List
	current       *list.Element
}

/*
* New() returns a new RoundRobin LoadBalancer
 */
func New(instances []skynet.ServiceInfo) loadbalancer.LoadBalancer {
	lb := &LoadBalancer{
		instances: make(map[string]*list.Element),
	}

	for _, i := range instances {
		lb.AddInstance(i)
	}

	return lb
}

func (lb *LoadBalancer) AddInstance(s skynet.ServiceInfo) {
	if _, ok := lb.instances[s.UUID]; ok {
		lb.UpdateInstance(s)
		return
	}

	lb.instanceMutex.Lock()
	defer lb.instanceMutex.Unlock()

	var e *list.Element

	if !s.Registered {
		e = &list.Element{Value: s}
	} else {
		e = lb.instanceList.PushBack(s)
	}

	lb.instances[s.UUID] = e
}

func (lb *LoadBalancer) UpdateInstance(s skynet.ServiceInfo) {
	if _, ok := lb.instances[s.UUID]; !ok {
		lb.AddInstance(s)
		return
	}

	lb.instanceMutex.Lock()
	defer lb.instanceMutex.Unlock()

	lb.instances[s.UUID].Value = s

	if s.Registered == false {
		lb.instanceList.Remove(lb.instances[s.UUID])
	}
}

func (lb *LoadBalancer) RemoveInstance(s skynet.ServiceInfo) {
	lb.instanceMutex.Lock()
	defer lb.instanceMutex.Unlock()

	lb.instanceList.Remove(lb.instances[s.UUID])
	delete(lb.instances, s.UUID)

	// current should be nil if we have no instances
	if lb.instanceList.Len() == 0 {
		lb.current = nil
	}
}

func (lb *LoadBalancer) Choose() (s skynet.ServiceInfo, err error) {
	if lb.current == nil {
		if lb.instanceList.Len() == 0 {
			return s, loadbalancer.NoInstances
		}

		lb.current = lb.instanceList.Front()
		return lb.current.Value.(skynet.ServiceInfo), nil
	}

	lb.current = lb.current.Next()

	if lb.current == nil {
		lb.current = lb.instanceList.Front()
	}

	s = lb.current.Value.(skynet.ServiceInfo)

	return s, nil
}
