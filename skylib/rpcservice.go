package skylib

import (
	"reflect"
	"rpc"
	"rpc/jsonrpc"
	"http"
	"net"
	"fmt"
	"log"
	"strings"
)

// Parent struct for the registry.
type RegisteredNetworkServers struct {
	Services []*RpcService
}

type ServerRegistry interface {
	Equal(that interface{}) bool
}


// This struct will be serialized and passed to Registry.
// If this is really a Server, then Provides should really be a list of
// all the Service classes provided by this Server.
type RpcService struct {
	IPAddress string
	Port      int
	Provides  string // Class name, not any specific method.
	Protocol  string // json, etc.
	l 		  net.Listener
}

func (r *RpcService) parseError(err string) {
	panic(&Error{err, r.Provides})
}

// At the moment, this can start more than one Server on the same
// port, which could be a problem.
func (self *RpcService) Serve(done chan bool) {


	switch self.Protocol {
	default:
		rpc.HandleHTTP() // Seems safe to call multiple times, but must
		// that precede net.Listen()?
		log.Println("Starting http server")
		http.Serve(self.l, nil)
	case "json":
		log.Println("Starting jsonrpc server")
		for {
			conn, err := self.l.Accept()
			if err != nil {
				panic(err.String())
			}
			jsonrpc.ServeConn(conn)
		}
	}
	done <- true // This may never occur.
}

func (this *RpcService) Equal(that *RpcService) bool {
	var b bool
	b = false
	if this.IPAddress != that.IPAddress {
		return b
	}
	if this.Port != that.Port {
		return b
	}
	if this.Provides != that.Provides {
		return b
	}
	if this.Protocol != that.Protocol {
		return b
	}
	b = true
	return b
}

func NewRpcService(sig interface{}) *RpcService {
	////star_name := reflect.TypeOf(sig).String())
	type_name := reflect.Indirect(reflect.ValueOf(sig)).Type().Name()
	rpc.Register(sig)
	r := &RpcService{
		Port:      *Port,
		IPAddress: *BindIP,
		Provides:  type_name,
		Protocol:  strings.ToLower(*Protocol),
	}

	portString := fmt.Sprintf("%s:%d", r.IPAddress, r.Port)
	log.Println(portString)

	l, e := net.Listen("tcp", portString)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	r.l = l
	t, e := net.ResolveTCPAddr("tcp",l.Addr().String())
	if e != nil {
		log.Fatal("listen error:", e)
	}
	r.Port = t.Port


	return r
}
