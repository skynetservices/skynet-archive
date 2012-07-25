package skylib

import (
	"code.google.com/p/gonicetrace/nicetrace"
	"errors"
	"fmt"
	"github.com/bketelsen/skynet/rpc/bsonrpc"
	"launchpad.net/mgo/v2/bson"
	"os"
	"reflect"
	"time"
)

var (
	RequestInfoPtrType = reflect.TypeOf(&RequestInfo{})

	anError   error
	ErrorType = reflect.TypeOf(&anError).Elem()
)

type ServiceRPC struct {
	log         Logger
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

func NewServiceRPC(sd ServiceDelegate, log Logger) (srpc *ServiceRPC) {
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

		// must have four parameters: (receiver, RequestInfo, something, something)
		if ftyp.NumIn() != 4 {
			goto problem
		}
		// don't have to check for the receiver
		if ftyp.In(1) != RequestInfoPtrType {
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

// ServiceRPC.Forward is the entry point for RPC calls
func (srpc *ServiceRPC) Forward(in ServiceRPCIn, out *ServiceRPCOut) (err error) {
	m, ok := srpc.methods[in.Method]
	if !ok {
		err = errors.New(fmt.Sprintf("No such method %q", in.Method))
		return
	}

	inValuePtr := reflect.New(m.Type().In(2))

	// fmt.Printf("in.In: %v\n", in.In)

	err = bsonrpc.CopyTo(in.In.(bson.M), inValuePtr.Interface())
	if err != nil {
		return
	}

	outValue := reflect.New(m.Type().In(3).Elem())
	// fmt.Printf("in: %T %v\n", inValuePtr.Elem().Interface(), inValuePtr.Elem().Interface())

	// fmt.Println("calling", in.Method)

	startTime := time.Now().UnixNano()

	returns := m.Call([]reflect.Value{
		reflect.ValueOf(srpc.delegate),
		reflect.ValueOf(in.RequestInfo),
		inValuePtr.Elem(),
		outValue,
	})

	duration := time.Now().UnixNano() - startTime

	mc := MethodCall{
		MethodName:  in.Method,
		RequestInfo: in.RequestInfo,
		Duration:    duration,
	}

	if srpc.log != nil {
		srpc.log.Item(mc)
	}

	out.Out = outValue.Elem().Interface()
	// fmt.Println("out:", out.Out)
	// fmt.Println("err:", out.Err)

	erri := returns[0].Interface()
	out.Err, _ = erri.(error)

	return
}

type ServiceRPCIn struct {
	Method      string
	RequestInfo *RequestInfo
	In          interface{}
}

type ServiceRPCOut struct {
	Out interface{}
	Err error
}

type RequestInfo struct {
	RequestID string
}
