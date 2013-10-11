package test

import (
	"errors"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/client/loadbalancer"
)

type LoadBalancer struct {
	AddInstanceFunc    func(s skynet.ServiceInfo)
	UpdateInstanceFunc func(s skynet.ServiceInfo)
	RemoveInstanceFunc func(s skynet.ServiceInfo)
	ChooseFunc         func() (skynet.ServiceInfo, error)
}

func NewLoadBalancer(instances []skynet.ServiceInfo) (lb loadbalancer.LoadBalancer) {
	return &LoadBalancer{}
}

func (lb *LoadBalancer) AddInstance(s skynet.ServiceInfo) {
	if lb.AddInstanceFunc != nil {
		lb.AddInstanceFunc(s)
	}
}

func (lb *LoadBalancer) UpdateInstance(s skynet.ServiceInfo) {
	if lb.UpdateInstanceFunc != nil {
		lb.UpdateInstanceFunc(s)
	}
}

func (lb *LoadBalancer) RemoveInstance(s skynet.ServiceInfo) {
	if lb.RemoveInstanceFunc != nil {
		lb.RemoveInstanceFunc(s)
	}
}

func (lb *LoadBalancer) Choose() (skynet.ServiceInfo, error) {
	if lb.ChooseFunc != nil {
		return lb.ChooseFunc()
	}

	return skynet.ServiceInfo{}, errors.New("No instances found that match that criteria")
}
