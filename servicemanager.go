package skynet

import (
	"github.com/skynetservices/skynet2/log"
)

const (
	_ = iota
	InstanceAdded
	InstanceRemoved
	InstanceUpdated
)

type InstanceNotification struct {
	Type    int
	Service ServiceInfo
}

type ServiceManager interface {
	Add(s ServiceInfo) error
	Update(s ServiceInfo) error
	Remove(s ServiceInfo) error
	Register(uuid string) error
	Unregister(uuid string) error

	Shutdown() error

	// Discovery
	ListHosts(c CriteriaMatcher) ([]string, error)
	ListRegions(c CriteriaMatcher) ([]string, error)
	ListServices(c CriteriaMatcher) ([]string, error)
	ListVersions(c CriteriaMatcher) ([]string, error)
	ListInstances(c CriteriaMatcher) ([]ServiceInfo, error)
	Watch(criteria CriteriaMatcher, c chan<- InstanceNotification) []ServiceInfo
}

var manager ServiceManager

func SetServiceManager(sm ServiceManager) {
	manager = sm
}

func GetServiceManager() ServiceManager {
	if manager == nil {
		log.Fatal("No ServiceManager provided")
	}

	return manager
}
