package client

import (
	"github.com/bketelsen/skynet"
	"sort"
	"time"
)

type InstanceComparator func(c *Client, i1, i2 *skynet.ServiceInfo) (i1IsLess bool)

const (
	SAME_REGION_POINTS            = 10
	SAME_HOST_POINTS              = 3
	REQUESTED_LAST_POINTS         = 1
	CRITICAL_CLIENTS_POINTS       = -15 //random number tbd
	CRITICAL_RESPONSE_TIME_POINTS = -17 //random number tbd
)

func getInstanceScore(c *Client, i *skynet.ServiceInfo) (points int) {
	if i.Config.ServiceAddr.IPAddress == c.Config.Host {
		points = points + SAME_HOST_POINTS
	}

	if i.Config.Region == c.Config.Region {
		points = points + SAME_REGION_POINTS
	}

	if i.Config.CriticalClientCount > 0 && i.Stats.Clients >= i.Config.CriticalClientCount {
		points = points + CRITICAL_CLIENTS_POINTS
	}

	if i.Config.CriticalAverageResponseTime > 0 && i.Stats.AverageResponseTime >= i.Config.CriticalAverageResponseTime {
		points = points + CRITICAL_RESPONSE_TIME_POINTS
	}

	return
}

func defaultComparator(c *Client, i1, i2 *skynet.ServiceInfo) bool {
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
	sort.Sort((*InstanceHeap)(ic))
	instance = ic.instances[0]
	return
}

type InstanceHeap InstanceChooser

func (h *InstanceHeap) Len() int {
	return len(h.instances)
}

func (h *InstanceHeap) Less(i, j int) bool {
	return h.comparator(h.client, h.instances[i], h.instances[j])
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
