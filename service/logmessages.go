package service

import (
	"fmt"
	"github.com/skynetservices/skynet2"
	"syscall"
)

type RegisteredMethods struct {
	Methods []string
}

func (rm RegisteredMethods) String() string {
	return fmt.Sprintf("Registered methods: %v", rm.Methods)
}

type MethodCall struct {
	RequestInfo *skynet.RequestInfo
	MethodName  string
}

func (mi MethodCall) String() string {
	return fmt.Sprintf("Method %q called with RequestInfo %v", mi.MethodName, mi.RequestInfo)
}

type MethodCompletion struct {
	RequestInfo *skynet.RequestInfo
	MethodName  string
	Duration    int64
}

func (mi MethodCompletion) String() string {
	return fmt.Sprintf("Method %q completed with RequestInfo %v and duration %dns", mi.MethodName, mi.RequestInfo, mi.Duration)
}

type KillSignal struct {
	Signal syscall.Signal
}

func (ks KillSignal) String() string {
	return fmt.Sprintf("Got kill signal %q", ks.Signal)
}

type ServiceListening struct {
	ServiceConfig *skynet.ServiceConfig
	Addr          *skynet.BindAddr
}

func (sc ServiceListening) String() string {
	return fmt.Sprintf("Service %q listening on %s", sc.ServiceConfig.Name, sc.Addr)
}

type ServiceRegistered struct {
	ServiceConfig *skynet.ServiceConfig
}

func (sr ServiceRegistered) String() string {
	return fmt.Sprintf("Service %q registered", sr.ServiceConfig.Name)
}

type ServiceUnregistered struct {
	ServiceConfig *skynet.ServiceConfig
}

func (sr ServiceUnregistered) String() string {
	return fmt.Sprintf("Service %q unregistered", sr.ServiceConfig.Name)
}
