//Copyright (c) 2011 Brian Ketelsen

//Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package main

import (
	"rpc"
	"log"
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


func main() {

	// Pull in command line options or defaults if none given
	flag.Parse()

	agent := skylib.NewAgent()
	agent.Start()

	go monitorServices()

	agent.Wait()
}
