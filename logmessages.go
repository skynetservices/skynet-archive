package skynet

import (
	"fmt"
)

type ServiceDiscovered struct {
	Service *ServiceInfo
}

func (sd ServiceDiscovered) String() string {
	return fmt.Sprintf("Discovered service %q at %s", sd.Service.Config.Name,
		sd.Service.Config.ServiceAddr)
}

type ServiceRemoved struct {
	Service *ServiceInfo
}

func (sr ServiceRemoved) String() string {
	return fmt.Sprintf("Removed service %q at %s", sr.Service.Config.Name,
		sr.Service.Config.ServiceAddr)
}

type ServiceCreated struct {
	ServiceConfig *ServiceConfig
}

func (sc ServiceCreated) String() string {
	return fmt.Sprintf("Created service %q", sc.ServiceConfig.Name)
}
