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
	"syscall"
)


func monitorServices() {
	for {
		skylib.LoadConfig()
		for _, v := range skylib.NS.Services {
			fmt.Println(v) 
			// Insert your code to do something for 
			// each service here
			// or get rid of this loop
			// and do something else interesting!
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

	skylib.Setup("Watcher.Generic")  // Change this to be more descriptive

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
