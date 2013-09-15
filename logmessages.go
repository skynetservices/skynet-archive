package skynet

import (
	"fmt"
)

type ServiceDiscovered struct {
	Service *ServiceInfo
}

func (sd ServiceDiscovered) String() string {
	return fmt.Sprintf("Discovered service %q at %s", sd.Service.Name,
		sd.Service.ServiceAddr)
}

type ServiceRemoved struct {
	Service *ServiceInfo
}

func (sr ServiceRemoved) String() string {
	return fmt.Sprintf("Removed service %q at %s", sr.Service.Name,
		sr.Service.ServiceAddr)
}

type ServiceCreated struct {
	ServiceInfo *ServiceInfo
}

func (sc ServiceCreated) String() string {
	return fmt.Sprintf("Created service %q", sc.ServiceInfo.Name)
}
