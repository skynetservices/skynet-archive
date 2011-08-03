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

// This struct will be serialized and passed to Config.
type RpcServer struct {
	IPAddress string
	Port      int
	Provides  string // Class name, not any specific method.
	Protocol  string // json, etc.
}

func (self *RpcServer) Serve(done chan bool) {
	portString := fmt.Sprintf("%s:%d", self.IPAddress, self.Port)
	log.Println(portString)

	l, e := net.Listen("tcp", portString)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	defer l.Close()

	switch self.Protocol {
	default:
		rpc.HandleHTTP() // Seems safe to call multiple times, but must
		// that precede net.Listen()?
		log.Println("Starting http server")
		http.Serve(l, nil)
	case "json":
		log.Println("Starting jsonrpc server")
		for {
			conn, err := l.Accept()
			if err != nil {
				panic(err.String())
			}
			jsonrpc.ServeConn(conn)
		}
	}
	done <- true
}

func (this *RpcServer) Equal(that *RpcServer) bool {
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

func NewRpcServer(model interface{}) *RpcServer {
	////star_name := reflect.TypeOf(model).String())
	type_name := reflect.Indirect(reflect.ValueOf(model)).Type().Name()
	rpc.Register(model)
	r := &RpcServer{
		Port:      *Port,
		IPAddress: *BindIP,
		Provides:  type_name,
		Protocol:  strings.ToLower(*Protocol),
	}
	return r
}
