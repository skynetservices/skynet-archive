package client

import (
	"container/heap"
	"fmt"
	"github.com/bketelsen/skynet"
	"time"
)

type InstanceComparator func(c *Client, i1, i2 *skynet.ServiceInfo) (i1IsLess bool)

const (
	SAME_REGION_POINTS         = 10
	SAME_HOST_POINTS           = 2
	REQUESTED_LAST_POINTS      = 1
	CRITICAL_NUMBER_OF_CLIENTS = 10 //random number tbd
	CRITICAL_RESPONSE_TIME     = 20 //random number tbd
)

func getInstanceScore(c *Client, i *skynet.ServiceInfo) (points int) {
	if i.Config.ServiceAddr.IPAddress == c.Config.Host {
		points = points + SAME_HOST_POINTS
	}

	if i.Config.Region == c.Config.Region {
		points = points + SAME_REGION_POINTS
	}

	return
}

func myComparator(c *Client, i1, i2 *skynet.ServiceInfo) (i1IsBetter bool) {
	var closer *skynet.ServiceInfo
	var far *skynet.ServiceInfo
	var i1Points = getInstanceScore(c, i1)
	var i2Points = getInstanceScore(c, i2)

	// All things being equal let's sort on LastRequest
	if i1Points == i2Points {
		var t1, t2 int64 = 0, 0

		t, err := time.Parse("2006-01-02T15:04:05Z-0700", i1.Stats.LastRequest)
		if err == nil {
			t1 = t.Unix()
		}

		t, err = time.Parse("2006-01-02T15:04:05Z-0700", i2.Stats.LastRequest)
		if err == nil {
			t2 = t.Unix()
		}

		if t1 < t2 {
			i1Points = i1Points + REQUESTED_LAST_POINTS
		} else {
			i2Points = i2Points + REQUESTED_LAST_POINTS
		}
	}

	if i1Points > i2Points {
		closer = i1
		far = i2
	} else {
		closer = i2
		far = i1
	}
	i1IsBetter = compareServers(closer, far)
	return i1IsBetter
}

func compareServers(closer, far *skynet.ServiceInfo) (closerIsBetter bool) {

	if closer.Stats.Clients > CRITICAL_NUMBER_OF_CLIENTS || closer.Stats.AverageResponseTime > CRITICAL_RESPONSE_TIME {
		if far.Stats.Clients <= CRITICAL_NUMBER_OF_CLIENTS && far.Stats.AverageResponseTime <= CRITICAL_RESPONSE_TIME {
			//chose far instance
			closerIsBetter = false
		} else { //we are in trouble, can not use both ?? 
			//TODO Figure out what to do - panic??
			panic("Both instances reached critical condition!")
		}
	} else {
		//chose closer instance
		closerIsBetter = true
	}
	return closerIsBetter
}

func defaultComparator(c *Client, i1, i2 *skynet.ServiceInfo) (i1IsBetter bool) {

	var i1Points = getInstanceScore(c, i1)
	var i2Points = getInstanceScore(c, i2)

	// TODO: Score Clients (make sure to account for 0 in case of new instances)
	// TODO: Score AverageResponseTime (make sure to account for 0 in case of new instances)

	// All things being equal let's sort on LastRequest
	if i1Points == i2Points {
		var t1, t2 int64 = 0, 0

		t, err := time.Parse("2006-01-02T15:04:05Z-0700", i1.Stats.LastRequest)
		if err == nil {
			t1 = t.Unix()
		}

		t, err = time.Parse("2006-01-02T15:04:05Z-0700", i2.Stats.LastRequest)
		if err == nil {
			t2 = t.Unix()
		}

		if t1 < t2 {
			i1Points = i1Points + REQUESTED_LAST_POINTS
		} else {
			i2Points = i2Points + REQUESTED_LAST_POINTS
		}
	}

	return i1Points > i2Points
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
		client:   c,
		addCh:    make(chan *skynet.ServiceInfo, 1),
		remCh:    make(chan *skynet.ServiceInfo, 1),
		chooseCh: make(chan chan *skynet.ServiceInfo),
	}

	if c.Config.Prioritizer != nil {
		ic.comparator = func(c *Client, i1, i2 *skynet.ServiceInfo) (i1IsBetter bool) {
			return c.Config.Prioritizer(i1, i2)
		}
	} else if c.Config.Region != skynet.DefaultRegion {
		ic.comparator = defaultComparator
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
	// if there is no comparator, choose randomly
	if ic.comparator == nil {
		i := ic.count % len(ic.instances)
		instance = ic.instances[i]
		ic.count++
		return
	}

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
