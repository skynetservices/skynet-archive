package skynet

import (
	"fmt"
)

type ServiceDiscovered struct {
	Service *ServiceInfo
}

func (sd ServiceDiscovered) String() string {
	return fmt.Sprintf("Discovered service %q at %s", sd.Service.ServiceConfig.Name,
		sd.Service.ServiceConfig.ServiceAddr)
}

type ServiceRemoved struct {
	Service *ServiceInfo
}

func (sr ServiceRemoved) String() string {
	return fmt.Sprintf("Removed service %q at %s", sr.Service.ServiceConfig.Name,
		sr.Service.ServiceConfig.ServiceAddr)
}

type ServiceCreated struct {
	ServiceConfig *ServiceConfig
}

func (sc ServiceCreated) String() string {
	return fmt.Sprintf("Created service %q", sc.ServiceConfig.Name)
}
