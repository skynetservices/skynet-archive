package service

import (
	"errors"
	"fmt"
	"github.com/skynetservices/skynet2"
	"github.com/skynetservices/skynet2/log"
	"github.com/skynetservices/skynet2/stats"
	"labix.org/v2/mgo/bson"
	"reflect"
	"time"
)

var (
	RequestInfoPtrType = reflect.TypeOf(&skynet.RequestInfo{})

	anError   error
	ErrorType = reflect.TypeOf(&anError).Elem()
)

type ServiceRPC struct {
	service     *Service
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

func NewServiceRPC(s *Service) (srpc *ServiceRPC) {
	srpc = &ServiceRPC{
		service: s,
		methods: make(map[string]reflect.Value),
	}

	// scan through methods looking for a method (RequestInfo,
	// something, something) error
	typ := reflect.TypeOf(srpc.service.Delegate)
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

		// must have four parameters: (receiver, RequestInfo,
		// somethingIn, somethingOut)
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
		log.Printf(log.WARN, "Bad RPC method for %T: %q %v\n", s.Delegate, m.Name, f)
	}

	return
}

// ServiceRPC.Forward is the entry point for RPC calls. It wraps actual RPC calls
// and provides a slot for the RequestInfo. The parameters to the actual RPC
// calls are transmitted in a []byte, and are then marshalled/unmarshalled on
// either end.
func (srpc *ServiceRPC) Forward(in skynet.ServiceRPCIn, out *skynet.ServiceRPCOut) (err error) {
	srpc.service.activeRequests.Add(1)
	defer srpc.service.activeRequests.Done()

	go stats.MethodCalled(in.Method)

	clientInfo, ok := srpc.service.getClientInfo(in.ClientID)
	if !ok {
		err = errors.New("did not provide the ClientID")
		log.Printf(log.ERROR, "%+v", MethodError{in.RequestInfo, in.Method, err})
		return
	}

	in.RequestInfo.ConnectionAddress = clientInfo.Address.String()
	if in.RequestInfo.OriginAddress == "" || !srpc.service.IsTrusted(clientInfo.Address) {
		in.RequestInfo.OriginAddress = in.RequestInfo.ConnectionAddress
	}

	mc := MethodCall{
		MethodName:  in.Method,
		RequestInfo: in.RequestInfo,
	}

	log.Printf(log.INFO, "%+v", mc)

	m, ok := srpc.methods[in.Method]
	if !ok {
		err = errors.New(fmt.Sprintf("No such method %q", in.Method))
		log.Printf(log.ERROR, "%+v", MethodError{in.RequestInfo, in.Method, err})
		return
	}

	inValuePtr := reflect.New(m.Type().In(2))

	err = bson.Unmarshal(in.In, inValuePtr.Interface())
	if err != nil {
		log.Println(log.ERROR, "Error unmarshaling request", err)
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
		err = errors.New("illegal out param type")
		log.Printf(log.ERROR, "%+v", MethodError{in.RequestInfo, in.Method, err})
		return
	}

	startTime := time.Now()

	params := []reflect.Value{
		reflect.ValueOf(srpc.service.Delegate),
		reflect.ValueOf(in.RequestInfo),
		inValuePtr.Elem(),
		outValue,
	}

	returns := m.Call(params)

	duration := time.Now().Sub(startTime)

	mcp := MethodCompletion{
		MethodName:  in.Method,
		RequestInfo: in.RequestInfo,
		Duration:    duration,
	}

	log.Printf(log.INFO, "%+v", mcp)

	out.Out, err = bson.Marshal(outValue.Interface())
	if err != nil {
		log.Printf(log.ERROR, "%+v", MethodError{in.RequestInfo, in.Method, fmt.Errorf("Error marshaling response", err)})
		return
	}

	var rerr error = nil
	erri := returns[0].Interface()
	if erri != nil {
		rerr, _ = erri.(error)
		out.ErrString = rerr.Error()

		log.Printf(log.ERROR, "%+v", MethodError{in.RequestInfo, in.Method, fmt.Errorf("Method returned error:", err)})
	}

	go stats.MethodCompleted(in.Method, duration, rerr)

	return
}
