package skynet

import (
	"github.com/skynetservices/skynet2/log"
)

type ServiceManager interface {
	Add(s ServiceInfo) error
	Update(s ServiceInfo) error
	Remove(s ServiceInfo) error
	Register(uuid string) error
	Unregister(uuid string) error

	// Discovery
	ListHosts(c Criteria) ([]string, error)
	ListRegions(c Criteria) ([]string, error)
	ListServices(c Criteria) ([]string, error)
	ListVersions(c Criteria) ([]string, error)
	ListInstances(c Criteria) ([]ServiceInfo, error)
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
