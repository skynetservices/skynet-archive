package skylib

import (
	//"reflect";
	"rpc"
	"rpc/jsonrpc"
	"http"
	"net"
	"fmt"
	"log"
)

type RpcServer struct {

}

func (*RpcServer) Serve() {
	portString := fmt.Sprintf("%s:%d", *BindIP, *Port)
	log.Println(portString)

	l, e := net.Listen("tcp", portString)
    if e != nil {
        log.Fatal("listen error:", e)
    }
    defer l.Close()

	switch *Protocol {
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
			go jsonrpc.ServeConn(conn)
		}
	}
}

func NewRpcServer(prov interface{}) *RpcServer {
	////star_name := reflect.TypeOf(prov).String())
	//sname := reflect.Indirect(reflect.ValueOf(prov)).Type().Name()
	rpc.Register(prov)
	return &RpcServer{}
}
