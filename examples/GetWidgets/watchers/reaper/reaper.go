//Copyright (c) 2011 Brian Ketelsen

//Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package main

import (
	"rpc"
	"os"
	"net"
	"log"
	"http"
	"flag"
	"github.com/bketelsen/skynet/skylib"
	"fmt"
	"time"
	"syscall"
)


func monitorServices() {
	for {
		skylib.LoadConfig()
		for _, v := range skylib.NS.Services {
			if (v.Port != *skylib.Port) || (v.IPAddress != *skylib.BindIP) {
				portString := fmt.Sprintf("%s:%d", v.IPAddress, v.Port)
				x, err := rpc.DialHTTP("tcp", portString)
				if err != nil {
					log.Println("BAD CON:", err)
					skylib.RemoveFromConfig(v)
					skylib.Errors.Add(1)
					break
				}
				hc := skylib.HeartbeatRequest{Timestamp: time.Seconds()}
				hcr := skylib.HeartbeatResponse{}
				err = x.Call("Service.Ping", hc, &hcr)
				if err != nil {
					log.Println(err.String())
					skylib.Errors.Add(1)
				}
				x.Close()
				skylib.Requests.Add(1)
			}
		}
		syscall.Sleep(2000 * 1000000) // sleep then do it again!
	}
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

	skylib.Setup("Watcher.Reaper")

	rpc.HandleHTTP()

	go monitorServices()

	portString := fmt.Sprintf("%s:%d", *skylib.BindIP, *skylib.Port)

	l, e := net.Listen("tcp", portString)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	log.Println("Starting server")
	http.Serve(l, nil)

}
