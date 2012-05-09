//Copyright (c) 2011 Brian Ketelsen

//Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package main

import "net/rpc"
import "os"
import "net"
import "log"
import "net/http"
import "github.com/bketelsen/skynet/skylib"
import "flag"
import "time"
import "fmt"

const sName = "GetUserDataService.GetUserData"

type GetUserDataRequest struct {
	YourInputValue string
}

type GetUserDataResponse struct {
	YourOutputValue string
	Errors          []string
}

type GetUserDataService struct {
	Version int
}

func NewGetUserDataService() *GetUserDataService {

	r := &GetUserDataService{
		Version: 1,
	}
	return r
}

func (ls *GetUserDataService) GetUserData(cr *GetUserDataRequest, lr *GetUserDataResponse) (err error) {
	result := make(chan string)
	timeout := make(chan bool)

	//This function produces the actual result
	go func() {
		time.Sleep(1e8) // force the fail
		result <- " was here"
	}()

	go func() {
		time.Sleep(1e9)
		timeout <- true
	}()

	select {
	case retVal := <-result:
		lr.YourOutputValue = cr.YourInputValue + retVal
	case <-timeout:
		lr.Errors = append(lr.Errors, "Service Timeout")
	}

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

	r := NewGetUserDataService()

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
