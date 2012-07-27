package service

import (
	"fmt"
	"github.com/bketelsen/skynet"
)

type ServiceDiscovered struct {
	Service *Service
}

func (sd ServiceDiscovered) String() string {
	return fmt.Sprintf("Discovered service %q at %s", sd.Service.Config.Name, sd.Service.Config.ServiceAddr)
}

type ServiceRemoved struct {
	Service *Service
}

func (sr ServiceRemoved) String() string {
	return fmt.Sprintf("Removed service %q at %s", sr.Service.Config.Name, sr.Service.Config.ServiceAddr)
}

type ServiceCreated struct {
	ServiceConfig *skynet.ServiceConfig
}

func (sc ServiceCreated) String() string {
	return fmt.Sprintf("Created service %q", sc.ServiceConfig.Name)
}

type ServiceListening struct {
	ServiceConfig *skynet.ServiceConfig
	Addr          *skynet.BindAddr
}

func (sc ServiceListening) String() string {
	return fmt.Sprintf("Service %q listening on %s", sc.ServiceConfig.Name, sc.Addr)
}

type AdminListening struct {
	ServiceConfig *skynet.ServiceConfig
}

func (al AdminListening) String() string {
	return fmt.Sprintf("Service %q listening for admin on %s", al.ServiceConfig.Name, al.ServiceConfig.AdminAddr)
}

type RegisteredMethods struct {
	Methods []string
}

func (rm RegisteredMethods) String() string {
	return fmt.Sprintf("Registered methods: %v", rm.Methods)
}

type MethodCall struct {
	RequestInfo *skynet.RequestInfo
	MethodName  string
	Duration    int64
}

func (mi MethodCall) String() string {
	return fmt.Sprintf("Method %q called with RequestInfo %v and duration %dns", mi.MethodName, mi.RequestInfo, mi.Duration)
}
