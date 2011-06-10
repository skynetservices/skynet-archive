package main

import "rpc"
import "os"
import "net"
import "log"
import "http"
import "github.com/bketelsen/skynet/skylib"
import "flag"
import "time"
import "fmt"
import "myStartup"

const sName = "GetUserDataService.GetUserData"

type GetUserDataService struct {
	Version int
}


func NewGetUserDataService() *GetUserDataService {

	r := &GetUserDataService{
		Version: 1,
	}
	return r
}




func (ls *GetUserDataService) GetUserData(cr *myStartup.GetUserDataRequest, lr *myStartup.GetUserDataResponse) (err os.Error) {
	result := make (chan string)
	timeout := make(chan bool)
	
	//This function produces the actual result
	go func() {
		time.Sleep(1e8) // force the fail
		result <- " was here"
	}()
	
	go func(){
		time.Sleep(1e9)
		timeout <- true
	}()
	
	
	select {
	case retVal := <-result:
			lr.YourOutputValue = cr.YourInputValue + retVal
	case <- timeout:
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
