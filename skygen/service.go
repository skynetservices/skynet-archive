package main

const serviceTemplate = `package main

import "rpc"
import "os"
import "net"
import "log"
import "http"
import "github.com/bketelsen/skynet/skylib"
import "flag"
import "fmt"
import "<%PackageName%>"

const sName = "<%ServiceName%>Service.<%ServiceName%>"

type <%ServiceName%>Service struct {
	Version int
}


func New<%ServiceName%>Service() *<%ServiceName%>Service {

	r := &<%ServiceName%>Service{
		Version: 1,
	}
	return r
}

func (ls *<%ServiceName%>Service) <%ServiceName%>(cr *<%PackageName%>.<%ServiceName%>Request, lr *<%PackageName%>.<%ServiceName%>Response) (err os.Error) {
	lr.YourOutputValue = "Hello World"
	skylib.Requests.Add(1)
	return nil
}


func main() {

	// Pull in command line options or defaults if none given
	flag.Parse()

	f, err := os.OpenFile(*skylib.LogFileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err == nil {
		defer f.Close()
		log.SetOutput(f)
	}

	skylib.Setup(sName)

	r := New<%ServiceName%>Service()

	rpc.Register(r)

	rpc.HandleHTTP()

	portString := fmt.Sprintf("%s:%d", *skylib.BindIP, *skylib.Port)

	l, e := net.Listen("tcp", portString)
	if e != nil {
		log.Fatal("listen error:", e)
	}

	log.Println("Starting server")
	http.Serve(l, nil)

}
`