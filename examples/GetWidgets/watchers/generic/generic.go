//Copyright (c) 2011 Brian Ketelsen

//Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package main

import (
	"flag"
	"github.com/bketelsen/skynet/skylib"
	"fmt"
	"syscall"
)


func monitorServices() {
	println("Fear the reaper...")
	for {
		skylib.LoadConfig()
		clients := skylib.GetAllClientsByService("CommonService")
		println("#Agents:", len(clients))
		for _, x := range clients {
			fmt.Printf("%v\n", x)
			// Insert your code to do something for 
			// each service here
			// or get rid of this loop
			// and do something else interesting!
			x.Close()
			skylib.Requests.Add(1) // Should we count Heartbeat requests?
		}
		syscall.Sleep(2000 * 1000000) // sleep then do it again!
	}
}


func main() {

	// Pull in command line options or defaults if none given
	flag.Parse()

	skylib.NewAgent().Start()

	monitorServices()
}
