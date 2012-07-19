package skylib

import (
	"errors"
	"fmt"
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

var reservedMethodNames = map[string]bool{}

func init() {

	var sd ServiceDelegate
	sdvalue := reflect.ValueOf(&sd).Elem().Type()
	for i := 0; i < sdvalue.NumMethod(); i++ {
		m := sdvalue.Method(i)
		reservedMethodNames[m.Name] = true
	}
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

		if reservedMethodNames[m.Name] {
			continue
		}

		f := m.Func
		ftyp := f.Type()

		// must have four parameters: (receiver, RequestInfo, something, something)
		if ftyp.NumIn() != 4 {
			goto problem
		}
		// don't have to check for the receiver
		if ftyp.In(1) != RequestInfoType {
			goto problem
		}

		// must have one return value that is an error
		if ftyp.NumOut() != 1 {
			goto problem
		}
		if ftyp.Out(0) != ErrorType {
			goto problem
		}

		// we've got a method!
		srpc.methods[m.Name] = f
		continue

	problem:
		panic(fmt.Sprintf("Bad RPC method for %T: %v", sd, f))
	}

	return
}

// ServiceRPC.Forward is the entry point for RPC calls
func (srpc *ServiceRPC) Forward(in ServiceRPCIn, out *ServiceRPCOut) (err error) {
	// TODO: something smart with panics? Or just let the server go down?
	/*
		defer func() {
			e := recover()
			if e != nil {
				// what?
			}
		}()
	*/

	m, ok := srpc.methods[in.Method]
	if !ok {
		err = errors.New(fmt.Sprintf("No such method %q", in.Method))
		return
	}

	outValue := reflect.New(m.Type().In(3).Elem())

	returns := m.Call([]reflect.Value{
		reflect.ValueOf(srpc.delegate),
		reflect.ValueOf(in.RequestInfo),
		reflect.ValueOf(in.In),
		outValue,
	})
	outReply := reflect.ValueOf(out.Out).Elem()
	outReply.Set(outValue.Elem())

	erri := returns[0].Interface()
	if erri != nil {
		out.Err = erri.(error)
	}

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
