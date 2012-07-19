package skylib

import (
	"reflect"
)

var (
	RequestInfoType = reflect.TypeOf(RequestInfo{})

	anError   error
	ErrorType = reflect.TypeOf(&anError).Elem()
)

type ServiceRPC struct {
	delegate ServiceDelegate
	methods  map[string]reflect.Value
}

func NewServiceRPC(sd ServiceDelegate) (srpc *ServiceRPC) {
	srpc = &ServiceRPC{
		delegate: sd,
		methods:  make(map[string]reflect.Value),
	}

	// scan through methods looking for a method (RequestInfo, something, something) error
	typ := reflect.TypeOf(srpc.delegate)
	for i := 0; i < typ.NumMethod(); i++ {
		m := typ.Method(i)
		f := m.Func
		ftyp := f.Type()

		// must have four parameters: (receiver, RequestInfo, something, something)
		if ftyp.NumIn() != 4 {
			continue
		}
		// don't have to check for the receiver
		if ftyp.In(1) != RequestInfoType {
			continue
		}

		// must have one return value that is an error
		if ftyp.NumOut() != 1 {
			continue
		}
		if ftyp.Out(0) != ErrorType {
			continue
		}

		// we've got a method!
		srpc.methods[m.Name] = f
	}

	return
}

// ServiceRPC.Forward is the entry point for RPC calls
func (srpc *ServiceRPC) Forward(in ServiceRPCIn, out *ServiceRPCOut) (err error) {
	return
}

type ServiceRPCIn struct {
	Method      string
	RequestInfo RequestInfo
	In          interface{}
}

type ServiceRPCOut struct {
	Out interface{}
	Err error
}

type RequestInfo struct {
	RequestID string
}
