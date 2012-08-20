package client

import (
	"github.com/bketelsen/skynet/service"
	"math/rand"
)

type InstanceChooser struct {
	instances []*service.Service

	addCh    chan *service.Service
	remCh    chan *service.Service
	chooseCh chan chan *service.Service
}

func NewInstanceChooser() (ic *InstanceChooser) {
	ic = &InstanceChooser{
		addCh:    make(chan *service.Service, 1),
		remCh:    make(chan *service.Service, 1),
		chooseCh: make(chan chan *service.Service),
	}

	go ic.mux()

	return
}

func (ic *InstanceChooser) mux() {
	var activeWaits []chan *service.Service
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

func (ic *InstanceChooser) Add(instance *service.Service) {
	ic.addCh <- instance
}

func (ic *InstanceChooser) add(instance *service.Service) {
	for _, in := range ic.instances {
		if in.GetConfigPath() == instance.GetConfigPath() {
			return
		}
	}
	ic.instances = append(ic.instances, instance)
}

func (ic *InstanceChooser) Remove(instance *service.Service) {
	ic.remCh <- instance
}

func (ic *InstanceChooser) remove(instance *service.Service) {
	for i, in := range ic.instances {
		if in.GetConfigPath() == instance.GetConfigPath() {
			ic.instances[i] = ic.instances[len(ic.instances)-1]
			ic.instances = ic.instances[:len(ic.instances)-1]
			return
		}
	}
}

func (ic *InstanceChooser) Choose(timeout chan bool) (instance *service.Service, ok bool) {
	ich := make(chan *service.Service, 1)
	ic.chooseCh <- ich
	select {
	case instance = <-ich:
		ok = true
	case <-timeout:
		ok = false
	}
	return
}

func (ic *InstanceChooser) choose() (instance *service.Service) {
	i := rand.Intn(len(ic.instances))
	instance = ic.instances[i]
	return
}
