package service

import (
	"fmt"
	"github.com/skynetservices/skynet"
	"syscall"
	"time"
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
	Duration    time.Duration
}

func (mi MethodCompletion) String() string {
	return fmt.Sprintf("Method %q completed with RequestInfo %v and duration %s", mi.MethodName, mi.RequestInfo, mi.Duration.String())
}

type MethodError struct {
	RequestInfo *skynet.RequestInfo
	MethodName  string
	Error       error
}

func (me MethodError) String() string {
	return fmt.Sprintf("Method %q failed with RequestInfo %v and error %s", me.MethodName, me.RequestInfo, me.Error.Error())
}

type KillSignal struct {
	Signal syscall.Signal
}

func (ks KillSignal) String() string {
	return fmt.Sprintf("Got kill signal %q", ks.Signal)
}

type ServiceListening struct {
	ServiceInfo *skynet.ServiceInfo
	Addr        *skynet.BindAddr
}

func (sc ServiceListening) String() string {
	return fmt.Sprintf("Service %q %q listening on %s in region %q", sc.ServiceInfo.Name, sc.ServiceInfo.Version, sc.Addr, sc.ServiceInfo.Region)
}

type ServiceRegistered struct {
	ServiceInfo *skynet.ServiceInfo
}

func (sr ServiceRegistered) String() string {
	return fmt.Sprintf("Service %q registered", sr.ServiceInfo.Name)
}

type ServiceUnregistered struct {
	ServiceInfo *skynet.ServiceInfo
}

func (sr ServiceUnregistered) String() string {
	return fmt.Sprintf("Service %q unregistered", sr.ServiceInfo.Name)
}
