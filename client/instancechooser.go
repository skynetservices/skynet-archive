package client

import (
	"container/heap"
	"github.com/bketelsen/skynet"
)

type InstanceComparator func(c *Client, i1, i2 *skynet.ServiceInfo) (i1IsLess bool)

func basicComparator(c *Client, i1, i2 *skynet.ServiceInfo) (i1IsBetter bool) {
	region := c.Config.Region
	// if only one has the right region, it's definitely less
	if region == i1.Config.Region && region != i2.Config.Region {
		return true
	}
	if region != i1.Config.Region && region == i2.Config.Region {
		return false
	}
	// otherwise use something arbitrary
	return i1.Config.UUID < i2.Config.UUID
}

type InstanceChooser struct {
	// client is whom we are choosing instances for
	client *Client

	comparator InstanceComparator

	instances []*skynet.ServiceInfo
	count     int

	addCh    chan *skynet.ServiceInfo
	remCh    chan *skynet.ServiceInfo
	chooseCh chan chan *skynet.ServiceInfo
}

func NewInstanceChooser(c *Client) (ic *InstanceChooser) {
	ic = &InstanceChooser{
		client:     c,
		comparator: basicComparator,
		addCh:      make(chan *skynet.ServiceInfo, 1),
		remCh:      make(chan *skynet.ServiceInfo, 1),
		chooseCh:   make(chan chan *skynet.ServiceInfo),
	}

	if c.Config.Prioritizer != nil {
		ic.comparator = func(c *Client, i1, i2 *skynet.ServiceInfo) (i1IsBetter bool) {
			return c.Config.Prioritizer(i1, i2)
		}
	}

	go ic.mux()

	return
}

func (ic *InstanceChooser) mux() {
	var activeWaits []chan *skynet.ServiceInfo
	for {
		select {
		case instance := <-ic.addCh:
			ic.add(instance)
			for _, ich := range activeWaits {
				ich <- instance
			}
			activeWaits = activeWaits[:0]
		case instance := <-ic.remCh:
			ic.remove(instance)
		case ich := <-ic.chooseCh:
			if len(ic.instances) == 0 {
				activeWaits = append(activeWaits, ich)
			} else {
				ich <- ic.choose()
			}
		}
	}
}

func (ic *InstanceChooser) Add(instance *skynet.ServiceInfo) {
	ic.addCh <- instance
}

func (ic *InstanceChooser) add(instance *skynet.ServiceInfo) {
	for _, in := range ic.instances {
		if in.GetConfigPath() == instance.GetConfigPath() {
			return
		}
	}
	ic.instances = append(ic.instances, instance)
}

func (ic *InstanceChooser) Remove(instance *skynet.ServiceInfo) {
	ic.remCh <- instance
}

func (ic *InstanceChooser) remove(instance *skynet.ServiceInfo) {
	for i, in := range ic.instances {
		if in.GetConfigPath() == instance.GetConfigPath() {
			ic.instances[i] = ic.instances[len(ic.instances)-1]
			ic.instances = ic.instances[:len(ic.instances)-1]
			return
		}
	}
}

func (ic *InstanceChooser) Choose(timeout chan bool) (instance *skynet.ServiceInfo, ok bool) {
	ich := make(chan *skynet.ServiceInfo, 1)
	ic.chooseCh <- ich
	select {
	case instance = <-ich:
		ok = true
	case <-timeout:
		ok = false
	}
	return
}

func (ic *InstanceChooser) choose() (instance *skynet.ServiceInfo) {
	// if the client's region is unknown, choose randomly
	if ic.client.Config.Region == skynet.DefaultRegion {
		i := ic.count % len(ic.instances) //rand.Intn(len(ic.instances))
		instance = ic.instances[i]
		ic.count++
		return
	}

	// otherwise, choose services in the same region

	// this heap-sorts (in linear time) the instances according to ic.comparator
	heap.Init((*InstanceHeap)(ic))
	instance = ic.instances[0]
	return
}

type InstanceHeap InstanceChooser

func (h *InstanceHeap) Len() int {
	return len(h.instances)
}

func (h *InstanceHeap) Less(i, j int) bool {
	// the indices are reversed here on purpose. if i<j, it goes at the end of the list rather than the beginning.
	return h.comparator(h.client, h.instances[j], h.instances[i])
}

func (h *InstanceHeap) Swap(i, j int) {
	h.instances[i], h.instances[j] = h.instances[j], h.instances[i]
}

func (h *InstanceHeap) Push(x interface{}) {
	panic("invalid use")
}

func (h *InstanceHeap) Pop() interface{} {
	panic("invalid use")
}
