package service

import (
	"code.google.com/p/gonicetrace/nicetrace"
	"errors"
	"fmt"
	"github.com/bketelsen/skynet"
	"launchpad.net/mgo/v2/bson"
	"os"
	"reflect"
	"time"
)

var (
	RequestInfoPtrType = reflect.TypeOf(&skynet.RequestInfo{})

	anError   error
	ErrorType = reflect.TypeOf(&anError).Elem()
)

type ServiceRPC struct {
	log         skynet.Logger
	delegate    ServiceDelegate
	methods     map[string]reflect.Value
	MethodNames []string
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

func NewServiceRPC(sd ServiceDelegate, log skynet.Logger) (srpc *ServiceRPC) {
	srpc = &ServiceRPC{
		log:      log,
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

		// this is the check to see if something is exported
		if m.PkgPath != "" {
			continue
		}

		f := m.Func
		ftyp := f.Type()

		// must have four parameters: (receiver, RequestInfo, somethingIn, somethingOut)
		if ftyp.NumIn() != 4 {
			goto problem
		}

		// don't have to check for the receiver

		// check the second parameter
		if ftyp.In(1) != RequestInfoPtrType {
			goto problem
		}

		// the somethingIn can be anything

		// somethingOut must be a pointer or a map
		switch ftyp.In(3).Kind() {
		case reflect.Ptr, reflect.Map:
		default:
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
		srpc.MethodNames = append(srpc.MethodNames, m.Name)
		continue

	problem:
		fmt.Println("trying to panic")
		fmt.Printf("Bad RPC method for %T: %q %v\n", sd, m.Name, f)
		nicetrace.WriteStacktrace(os.Stdout)
		panic(fmt.Sprintf("Bad RPC method for %T: %q %v\n", sd, m.Name, f))
	}

	return
}

// ServiceRPC.Forward is the entry point for RPC calls. It wraps actual RPC calls
// and provides a slot for the RequestInfo. The parameters to the actual RPC
// calls are transmitted in a []byte, and are then marshalled/unmarshalled on
// either end.
func (srpc *ServiceRPC) Forward(in ServiceRPCIn, out *ServiceRPCOut) (err error) {
	m, ok := srpc.methods[in.Method]
	if !ok {
		err = errors.New(fmt.Sprintf("No such method %q", in.Method))
		return
	}

	inValuePtr := reflect.New(m.Type().In(2))

	err = bson.Unmarshal(in.In, inValuePtr.Interface())
	if err != nil {
		return
	}

	// Allocate the out parameter of the RPC call.
	outType := m.Type().In(3)
	var outValue reflect.Value

	switch outType.Kind() {
	case reflect.Ptr:
		outValue = reflect.New(m.Type().In(3).Elem())
	case reflect.Map:
		outValue = reflect.MakeMap(outType)
	default:
		panic("illegal out param type")
	}

	startTime := time.Now().UnixNano()

	params := []reflect.Value{
		reflect.ValueOf(srpc.delegate),
		reflect.ValueOf(in.RequestInfo),
		inValuePtr.Elem(),
		outValue,
	}

	returns := m.Call(params)

	duration := time.Now().UnixNano() - startTime

	mc := MethodCall{
		MethodName:  in.Method,
		RequestInfo: in.RequestInfo,
		Duration:    duration,
	}

	if srpc.log != nil {
		srpc.log.Item(mc)
	}

	out.Out, err = bson.Marshal(outValue.Interface())
	if err != nil {
		return
	}

	erri := returns[0].Interface()
	out.Err, _ = erri.(error)

	return
}

type ServiceRPCIn struct {
	Method      string
	RequestInfo *skynet.RequestInfo
	In          []byte
}

type ServiceRPCOut struct {
	Out []byte
	Err error
}
