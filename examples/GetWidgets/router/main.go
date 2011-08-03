//Copyright (c) 2011 Brian Ketelsen

//Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package main

import (
	"os"
	"log"
	"flag"
	"time"
	"container/vector"
	"json"
	"github.com/bketelsen/skynet/skylib"

)

var route *skylib.Route

//const sName = "RouteService.RouteGetUserDataRequest"
const sName = "RouteService"

//Exporter struct for RPC
type RouteService struct {
	Name string
}


func callRpcService(model string, operation string, async bool, failOnErr bool, cr *skylib.SkynetRequest, rep *skylib.SkynetResponse) (err os.Error) {
	defer skylib.CheckError(&err)

	rpcClient, err := skylib.GetRandomClientByProvides(model)
	if err != nil {
		log.Println("No server provides", model)
		if failOnErr {
			return skylib.NewError(skylib.NO_CLIENT_PROVIDES_SERVICE, sName)
		} else {
			return nil
		}
	}
	name := model + operation
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
		rpcClient, err := skylib.GetRandomClientByProvides(model)
		err = rpcClient.Call(name, cr, rep)
		if err != nil {
			return skylib.NewError(err.String(), sName)
		}
	}
	log.Println("Called service sync", name)
	return nil
}


func (rs *RouteService) RouteGetUserDataRequest(cr *skylib.SkynetRequest, rep *skylib.SkynetResponse) (err os.Error) {
	defer skylib.CheckError(&err)
	log.Println(route)
	for i := 0; i < route.RouteList.Len(); i++ {
		rpcCall := route.RouteList.At(i).(map[string]interface{})

		err := callRpcService(rpcCall["Model"].(string), rpcCall["Method"].(string), rpcCall["Async"].(bool), rpcCall["ErrOnFail"].(bool), cr, rep)
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

	agent := skylib.NewAgent()
	agent.Start()

	CreateInitialRoute()

	route, err = skylib.GetRoute(sName)
	if err != nil {
		log.Panic("Unable to retrieve route.")
	}

	r := &RouteService{Name: *skylib.Name}
	agent.RegisterRpcServer(r)
	agent.Wait()
}

// Today this function creates a route in Doozer for the
// RouteService.RouteCreditRequest operation - which is CLARITY SPECIFIC
// and adds it too Doozer
func CreateInitialRoute() {

	// Create a basic Route object.
	r := &skylib.Route{}
	r.Name = sName
	r.LastUpdated = time.Seconds()
	r.Revision = 1
	r.RouteList = new(vector.Vector)

	// Define the chain of services.
	rpcScore := &skylib.RpcCall{
		Model: "GetUserDataService",
		Method: ".GetUserData",
		Async: false,
		OkToRetry: false,
		ErrOnFail: true,
	}

	// Just one, for now.
	r.RouteList.Push(rpcScore)

	// Marshal the route object into JSON.
	b, err := json.Marshal(r)
	if err != nil {
		log.Panic(err.String())
	}

	// Get the current revision number of the doozer "store".
	rev, err := skylib.DC.Rev()
	if err != nil {
		log.Panic(err.String())
	}

	// Set the contents of a file in the "store" to our JSON
	// string, if the file has not been modified since rev.
	filename := "/routes/" + sName
	_, err = skylib.DC.Set(filename, rev, b)
	if err != nil {
		log.Panic(err.String())
	}
	return
}
