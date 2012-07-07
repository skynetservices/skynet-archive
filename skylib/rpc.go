package skylib

/*
 * Modified version of msgpack-rpc/go
 *
 * this will be modified/refactored, or rewritten later but currently being used to test a theory
 *
 */

import (
	"errors"
	"fmt"
	msgpack "github.com/msgpack/msgpack-go"
	"io"
	"log"
	"net"
	"os"
	"reflect"
)

const (
	REQUEST      = 0
	RESPONSE     = 1
	NOTIFICATION = 2
)

type FunctionResolver interface {
	Resolve(name string, arguments []reflect.Value) (interface{}, reflect.Value, error)
}

type RpcServer struct {
	resolver     FunctionResolver
	log          *log.Logger
	listeners    []net.Listener
	autoCoercing bool
	lchan        chan int
}

// Goes into the event loop to get ready to serve.
func (self *RpcServer) Run() *RpcServer {
	lchan := make(chan int)
	for _, listener := range self.listeners {
		go (func(listener net.Listener) {
			for {
				conn, err := listener.Accept()
				if err != nil {
					self.log.Println(err)
					continue
				}
				if self.lchan == nil {
					conn.Close()
					break
				}
				go (func() {
				NextRequest:
					for {
						data, _, err := msgpack.UnpackReflected(conn)
						if err == io.EOF {
							break
						} else if err != nil {
							self.log.Println(err)
							break
						}
						msgId, funcName, _arguments, xerr := HandleRPCRequest(data)
						if xerr != nil {
							self.log.Println(xerr)
							continue NextRequest
						}

						service, f, xerr := self.resolver.Resolve(funcName, _arguments)
						if xerr != nil {
							self.log.Println(xerr)
							SendErrorResponseMessage(conn, msgId, xerr.Error())
							continue NextRequest
						}

						funcType := f.Type()
						if funcType.NumIn()-1 != len(_arguments) {
							msg := fmt.Sprintf("The number of the given arguments (%d) doesn't match the arity (%d)", len(_arguments), funcType.NumIn())
							self.log.Println(msg)
							SendErrorResponseMessage(conn, msgId, msg)
							continue NextRequest
						}

						if funcType.NumOut() != 1 && funcType.NumOut() != 2 {
							self.log.Println("The number of return values must be 1 or 2")
							SendErrorResponseMessage(conn, msgId, "Internal server error")
							continue NextRequest
						}

						arguments := make([]reflect.Value, funcType.NumIn())
						arguments[0] = reflect.ValueOf(service)

						for i, v := range _arguments {
							key := i + 1

							ft := funcType.In(key)
							vt := v.Type()

							if vt.AssignableTo(ft) {
								arguments[key] = v
							} else if pv, ok := integerPromote(ft, v); ok {
								arguments[key] = pv
							} else if self.autoCoercing && ft != nil && ft.Kind() == reflect.String && (v.Type().Kind() == reflect.Array || v.Type().Kind() == reflect.Slice) && (v.Type().Elem().Kind() == reflect.Uint8) {
								arguments[key] = reflect.ValueOf(string(v.Interface().([]byte)))
							} else {
								msg := fmt.Sprintf("The type of argument #%d doesn't match (%s expected, got %s)", i, ft.String(), vt.String())
								self.log.Println(msg)
								SendErrorResponseMessage(conn, msgId, msg)
								continue NextRequest
							}
						}

						retvals := f.Call(arguments)
						if funcType.NumOut() == 1 {
							SendResponseMessage(conn, msgId, retvals[0])
							continue NextRequest
						}
						var errMsg fmt.Stringer = nil
						_errMsg := retvals[1].Interface()
						if _errMsg != nil {
							var ok bool
							errMsg, ok = _errMsg.(fmt.Stringer)
							if !ok {
								self.log.Println("The second argument must have an interface { String() string }")
								SendErrorResponseMessage(conn, msgId, "Internal server error")
								continue NextRequest
							}
						}
						if errMsg != nil {
							SendErrorResponseMessage(conn, msgId, errMsg.String())
							continue NextRequest
						}
						if self.autoCoercing {
							_retval := retvals[0]
							if _retval.Kind() == reflect.String {
								retvals[0] = reflect.ValueOf([]byte(_retval.String()))
							}
						}
						SendResponseMessage(conn, msgId, retvals[0])
					}
					conn.Close()
				})()
			}
		})(listener)
	}
	self.lchan = lchan
	<-lchan
	for _, listener := range self.listeners {
		listener.Close()
	}
	return self
}

// integerPromote determines if we can promote v to dType, and if so, return the promoted value.
// This is needed because msgpack always encodes values as the minimum sized int that can hold them.
func integerPromote(dType reflect.Type, v reflect.Value) (reflect.Value, bool) {

	vt := v.Type()
	dsz := dType.Size()
	vtsz := vt.Size()

	if isIntType(dType) && isIntType(vt) && vtsz <= dsz {
		pv := reflect.New(dType).Elem()
		pv.SetInt(v.Int())
		return pv, true
	}

	if isUintType(dType) && isUintType(vt) && vtsz <= dsz {
		pv := reflect.New(dType).Elem()
		pv.SetUint(v.Uint())
		return pv, true
	}

	if isIntType(dType) && isUintType(vt) && vtsz <= dsz {
		pv := reflect.New(dType).Elem()
		pv.SetInt(int64(v.Uint()))
		return pv, true
	}

	if isUintType(dType) && isIntType(vt) && vtsz <= dsz {
		pv := reflect.New(dType).Elem()
		pv.SetUint(uint64(v.Int()))
		return pv, true
	}

	return v, false
}

type kinder interface {
	Kind() reflect.Kind
}

func isIntType(t kinder) bool {
	return t.Kind() == reflect.Int ||
		t.Kind() == reflect.Int8 ||
		t.Kind() == reflect.Int16 ||
		t.Kind() == reflect.Int32 ||
		t.Kind() == reflect.Int64
}

func isUintType(t kinder) bool {
	return t.Kind() == reflect.Uint ||
		t.Kind() == reflect.Uint8 ||
		t.Kind() == reflect.Uint16 ||
		t.Kind() == reflect.Uint32 ||
		t.Kind() == reflect.Uint64
}

// Lets the server quit the event loop
func (self *RpcServer) Stop() *RpcServer {
	if self.lchan != nil {
		lchan := self.lchan
		self.lchan = nil
		lchan <- 1
	}
	return self
}

// Listenes on the specified transport.  A single server can listen on the
// multiple ports.
func (self *RpcServer) Listen(listener net.Listener) *RpcServer {
	self.listeners = append(self.listeners, listener)
	return self
}

// Creates a new Server instance. raw bytesc are automatically converted into
// strings if autoCoercing is enabled.
func NewRpcServer(resolver FunctionResolver, autoCoercing bool, _log *log.Logger) *RpcServer {
	if _log == nil {
		_log = log.New(os.Stderr, "msgpack: ", log.Ldate|log.Ltime)
	}
	return &RpcServer{resolver, _log, make([]net.Listener, 0), autoCoercing, nil}
}

// This is a low-level function that is not supposed to be called directly
// by the user.  Change this if the MessagePack protocol is updated.
func HandleRPCRequest(req reflect.Value) (uint, string, []reflect.Value, error) {
	for {
		_req, ok := req.Interface().([]reflect.Value)
		if !ok {
			break
		}
		if len(_req) != 4 {
			break
		}
		msgType := _req[0]
		typeOk := msgType.Kind() == reflect.Int || msgType.Kind() == reflect.Int8 || msgType.Kind() == reflect.Int16 || msgType.Kind() == reflect.Int32 || msgType.Kind() == reflect.Int64
		if !typeOk {
			break
		}
		msgId := _req[1]
		idOk := msgId.Kind() == reflect.Int || msgId.Kind() == reflect.Int8 || msgId.Kind() == reflect.Int16 || msgId.Kind() == reflect.Int32 || msgId.Kind() == reflect.Int64 || msgId.Kind() == reflect.Uint8 || msgId.Kind() == reflect.Uint16 || msgId.Kind() == reflect.Uint32 || msgId.Kind() == reflect.Uint64
		if !idOk {
			break
		}
		_funcName := _req[2]
		funcOk := _funcName.Kind() == reflect.Array || _funcName.Kind() == reflect.Slice
		if !funcOk {
			break
		}
		funcName, ok := _funcName.Interface().([]uint8)
		if !ok {
			break
		}
		if msgType.Int() != REQUEST {
			break
		}
		_arguments := _req[3]
		var arguments []reflect.Value
		if _arguments.Kind() == reflect.Array || _arguments.Kind() == reflect.Slice {
			elemType := _req[3].Type().Elem()
			_elemType := elemType
			ok := _elemType.Kind() == reflect.Uint || _elemType.Kind() == reflect.Uint8 || _elemType.Kind() == reflect.Uint16 || _elemType.Kind() == reflect.Uint32 || _elemType.Kind() == reflect.Uint64 || _elemType.Kind() == reflect.Uintptr
			if !ok || _elemType.Kind() != reflect.Uint8 {
				arguments, ok = _arguments.Interface().([]reflect.Value)
			} else {
				arguments = []reflect.Value{reflect.ValueOf(string(_req[3].Interface().([]byte)))}
			}
		} else {
			arguments = []reflect.Value{_req[3]}
		}

		// Message Id could be an int or a uint (always > 0 though so we can cast it to uint)
		var id uint = 0

		if msgId.Kind() == reflect.Int || msgId.Kind() == reflect.Int8 || msgId.Kind() == reflect.Int16 || msgId.Kind() == reflect.Int32 || msgId.Kind() == reflect.Int64 {
			id = uint(msgId.Int())
		} else if msgId.Kind() == reflect.Uint8 || msgId.Kind() == reflect.Uint16 || msgId.Kind() == reflect.Uint32 || msgId.Kind() == reflect.Uint64 {
			id = uint(msgId.Uint())
		}

		return id, string(funcName), arguments, nil
	}
	return 0, "", nil, errors.New("Invalid message format")
}

// This is a low-level function that is not supposed to be called directly
// by the user.  Change this if the MessagePack protocol is updated.
func SendResponseMessage(writer io.Writer, msgId uint, value reflect.Value) error {
	_, err := writer.Write([]byte{0x94})
	if err != nil {
		return err
	}
	_, err = msgpack.PackInt8(writer, RESPONSE)
	if err != nil {
		return err
	}
	_, err = msgpack.PackUint(writer, msgId)
	if err != nil {
		return err
	}
	_, err = msgpack.PackNil(writer)
	if err != nil {
		return err
	}
	_, err = msgpack.PackValue(writer, value)
	return err
}

// This is a low-level function that is not supposed to be called directly
// by the user.  Change this if the MessagePack protocol is updated.
func SendErrorResponseMessage(writer io.Writer, msgId uint, errMsg string) error {
	_, err := writer.Write([]byte{0x94})
	if err != nil {
		return err
	}
	_, err = msgpack.PackInt8(writer, RESPONSE)
	if err != nil {
		return err
	}
	_, err = msgpack.PackUint(writer, msgId)
	if err != nil {
		return err
	}
	_, err = msgpack.PackBytes(writer, []byte(errMsg))
	if err != nil {
		return err
	}
	_, err = msgpack.PackNil(writer)
	return err
}
