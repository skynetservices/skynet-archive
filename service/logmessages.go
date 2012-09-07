package service

import (
	"fmt"
	"github.com/bketelsen/skynet"
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
