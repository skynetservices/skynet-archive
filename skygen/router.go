package main

const routerTemplate = `package main

import (
	"rpc"
	"os"
	"net"
	"log"
	"http"
	"flag"
	"github.com/bketelsen/skynet/skylib"
	"time"
	"container/vector"
	"json"
	"fmt"
	"<%PackageName%>"
)


var route *skylib.Route

const sName = "RouteService.Route<%ServiceName%>Request"

//Exporter struct for RPC
type RouteService struct {
	Name string
}


func callRpcService(name string, async bool, failOnErr bool, cr *<%PackageName%>.<%ServiceName%>Request, rep *<%PackageName%>.<%ServiceName%>Response) (err os.Error) {
	defer checkError(&err)

	rpcClient, err := skylib.GetRandomClientByProvides(name)
	if err != nil {
		log.Println("No service provides", name)
		if failOnErr {
			return skylib.NewError(skylib.NO_CLIENT_PROVIDES_SERVICE, sName)
		} else {
			return nil
		}
	}
	if async {
		go rpcClient.Call(name, cr, rep)
		log.Println("Called service async", name)
		return nil
	}
	log.Println("Calling : " + name)
	err = rpcClient.Call(name, cr, rep)
	if err != nil {
		log.Println("failed connection, retrying", err)
		// get another one and try again!
		rpcClient, err := skylib.GetRandomClientByProvides(name)
		err = rpcClient.Call(name, cr, rep)
		if err != nil {
			return skylib.NewError(err.String(), sName)
		}
	}
	log.Println("Called service sync", name)
	return nil
}


func (rs *RouteService) Route<%ServiceName%>Request(cr *<%PackageName%>.<%ServiceName%>Request, rep *<%PackageName%>.<%ServiceName%>Response) (err os.Error) {
	defer checkError(&err)
	log.Println(route)
	for i := 0; i < route.RouteList.Len(); i++ {
		rpcCall := route.RouteList.At(i).(map[string]interface{})

		err := callRpcService(rpcCall["Service"].(string), rpcCall["Async"].(bool), rpcCall["ErrOnFail"].(bool), cr, rep)
		if err != nil {
			skylib.Errors.Add(1)
			return err
		}

	}

	skylib.Requests.Add(1)
	return nil
}


// The Router application registers RPC listeners to accept from the initiators
// then registers RPC clients to each of the external services it may call.
func main() {

	var err os.Error

	// Pull in command line options or defaults if none given
	flag.Parse()

	f, err := os.OpenFile(*skylib.LogFileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err == nil {
		defer f.Close()
		log.SetOutput(f)
	}

	skylib.Setup(sName)

	route, err = skylib.GetRoute(sName)
	if err != nil {
		CreateInitialRoute()
	}

	r := &RouteService{Name: *skylib.Name}

	skylib.RegisterHeartbeat()
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

// checkError is a deferred function to turn a panic with type *Error into a plain error return.
// Other panics are unexpected and so are re-enabled.
func checkError(error *os.Error) {
	if v := recover(); v != nil {
		if e, ok := v.(*skylib.Error); ok {
			*error = e
		} else {
			// runtime errors should crash
			panic(v)
		}
	}
}
// Today this function creates a route in Doozer for the
// RouteService.RouteCreditRequest method - which is CLARITY SPECIFIC
// and adds it too Doozer
func CreateInitialRoute() (r skylib.Route) {

	r = skylib.Route{}
	r.Name = "RouteService.Route<%ServiceName%>Request"
	r.LastUpdated = time.Seconds()
	r.Revision = 1

	rpcScore := &skylib.RpcCall{Service: "<%ServiceName%>Service.<%ServiceName%>", Async: false, OkToRetry: false, ErrOnFail: true}

	rl := new(vector.Vector)

	r.RouteList = rl
	rl.Push(rpcScore)

	b, err := json.Marshal(r)
	if err != nil {
		log.Panic(err.String())
	}
	rev, err := skylib.DC.Rev()
	if err != nil {
		log.Panic(err.String())
	}
	_, err = skylib.DC.Set("/routes/RouteService.RouteGetACHDataRequest", rev, b)
	if err != nil {
		log.Panic(err.String())
	}
	return
}
`