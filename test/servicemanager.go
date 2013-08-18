package test

import (
	"github.com/skynetservices/skynet2"
)

type ServiceManager struct {
	AddFunc        func(s skynet.ServiceInfo) error
	UpdateFunc     func(s skynet.ServiceInfo) error
	RemoveFunc     func(s skynet.ServiceInfo) error
	RegisterFunc   func(uuid string) error
	UnregisterFunc func(uuid string) error

	ShutdownFunc func() error

	ListHostsFunc     func(c skynet.CriteriaMatcher) ([]string, error)
	ListRegionsFunc   func(c skynet.CriteriaMatcher) ([]string, error)
	ListServicesFunc  func(c skynet.CriteriaMatcher) ([]string, error)
	ListVersionsFunc  func(c skynet.CriteriaMatcher) ([]string, error)
	ListInstancesFunc func(c skynet.CriteriaMatcher) ([]skynet.ServiceInfo, error)
	WatchFunc         func(criteria skynet.CriteriaMatcher, c <-chan skynet.InstanceNotification) []skynet.ServiceInfo
}

func (sm *ServiceManager) Add(s skynet.ServiceInfo) error {
	if sm.AddFunc != nil {
		return sm.AddFunc(s)
	}

	return nil
}

func (sm *ServiceManager) Update(s skynet.ServiceInfo) error {
	if sm.UpdateFunc != nil {
		return sm.UpdateFunc(s)
	}

	return nil
}

func (sm *ServiceManager) Remove(s skynet.ServiceInfo) error {
	if sm.RemoveFunc != nil {
		return sm.RemoveFunc(s)
	}

	return nil
}

func (sm *ServiceManager) Register(uuid string) error {
	if sm.RegisterFunc != nil {
		return sm.RegisterFunc(uuid)
	}

	return nil
}

func (sm *ServiceManager) Unregister(uuid string) error {
	if sm.UnregisterFunc != nil {
		return sm.UnregisterFunc(uuid)
	}

	return nil
}

func (sm *ServiceManager) Shutdown() error {
	if sm.ShutdownFunc != nil {
		return sm.ShutdownFunc()
	}

	return nil
}

func (sm *ServiceManager) ListHosts(c skynet.CriteriaMatcher) ([]string, error) {
	if sm.ListHostsFunc != nil {
		return sm.ListHostsFunc(c)
	}

	return []string{}, nil
}

func (sm *ServiceManager) ListRegions(c skynet.CriteriaMatcher) ([]string, error) {
	if sm.ListRegionsFunc != nil {
		return sm.ListRegionsFunc(c)
	}

	return []string{}, nil
}

func (sm *ServiceManager) ListServices(c skynet.CriteriaMatcher) ([]string, error) {
	if sm.ListServicesFunc != nil {
		return sm.ListServicesFunc(c)
	}

	return []string{}, nil
}

func (sm *ServiceManager) ListVersions(c skynet.CriteriaMatcher) ([]string, error) {
	if sm.ListVersionsFunc != nil {
		return sm.ListVersionsFunc(c)
	}

	return []string{}, nil
}

func (sm *ServiceManager) ListInstances(c skynet.CriteriaMatcher) ([]skynet.ServiceInfo, error) {
	if sm.ListInstancesFunc != nil {
		return sm.ListInstancesFunc(c)
	}

	return []skynet.ServiceInfo{}, nil
}

func (sm *ServiceManager) Watch(criteria skynet.CriteriaMatcher, c <-chan skynet.InstanceNotification) (s []skynet.ServiceInfo) {
	if sm.ListInstancesFunc != nil {
		return sm.WatchFunc(criteria, c)
	}

	return
}
