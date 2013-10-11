package loadbalancer

import (
	"errors"
	"github.com/skynetservices/skynet"
)

var (
	NoInstances = errors.New("No instances")
)

type LoadBalancer interface {
	AddInstance(s skynet.ServiceInfo)
	UpdateInstance(s skynet.ServiceInfo)
	RemoveInstance(s skynet.ServiceInfo)
	Choose() (skynet.ServiceInfo, error)
}

type Factory func(instances []skynet.ServiceInfo) LoadBalancer
