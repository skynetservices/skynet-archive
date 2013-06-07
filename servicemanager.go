package skynet

import (
	"github.com/skynetservices/skynet2/log"
)

type ServiceManager interface {
	Add(s ServiceInfo)
	Update(s ServiceInfo)
	Remove(uuid string)
	Register(uuid string)
	Unregister(uuid string)
	ListRegions(query ServiceQuery) []string
	ListServices(query ServiceQuery) []string
	ListInstances(query ServiceQuery) []ServiceInfo
	ListHosts(query ServiceQuery) []string
	Subscribe(query ServiceQuery) chan ServiceUpdate
}

type ServiceQuery struct {
	UUID        []string
	Name        []string
	Version     []string
	Region      []string
	ServiceAddr []*BindAddr
}

type ServiceStatus int

const (
	ADD ServiceStatus = iota
	REMOVE
	UPDATE
)

type ServiceUpdate struct {
	Service ServiceInfo
	Event   ServiceStatus
}

var manager ServiceManager

func SetServiceManager(sm ServiceManager) {
	manager = sm
}

func GetServiceManager() ServiceManager {
	if manager == nil {
		log.Panic("No ServiceManager provided")
	}

	return manager
}
