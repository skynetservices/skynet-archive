//Copyright (c) 2011 Brian Ketelsen

//Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
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