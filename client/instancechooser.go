package client

import (
	"github.com/bketelsen/skynet"
	//"math/rand"
)

type InstanceChooser struct {
	instances []*skynet.ServiceInfo
	count     int

	addCh    chan *skynet.ServiceInfo
	remCh    chan *skynet.ServiceInfo
	chooseCh chan chan *skynet.ServiceInfo
}

func NewInstanceChooser() (ic *InstanceChooser) {
	ic = &InstanceChooser{
		addCh:    make(chan *skynet.ServiceInfo, 1),
		remCh:    make(chan *skynet.ServiceInfo, 1),
		chooseCh: make(chan chan *skynet.ServiceInfo),
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
	i := ic.count % len(ic.instances) //rand.Intn(len(ic.instances))
	instance = ic.instances[i]
	ic.count++
	return
}
