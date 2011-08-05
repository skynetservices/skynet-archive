//Copyright (c) 2011 Brian Ketelsen

//Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package skylib

import (
	"log"
	"os"
	"expvar"
	"syscall"
	"os/signal"
)


func initLogging() {
	f, err := os.OpenFile(*LogFileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err == nil {
		defer f.Close()
		log.SetOutput(f)
	}
}

func initDefaultExpVars(name string) {
	Requests = expvar.NewInt(name + "-processed")
	Errors = expvar.NewInt(name + "-errors")
	Goroutines = expvar.NewInt(name + "-goroutines")
}

func watchSignals() {

	for {
		select {
		case sig := <-signal.Incoming:
			switch sig.(os.UnixSignal) {
			case syscall.SIGUSR1:
				*LogLevel = *LogLevel + 1
				LogError("Loglevel changed to : ", *LogLevel)

			case syscall.SIGUSR2:
				if *LogLevel > 1 {
					*LogLevel = *LogLevel - 1
				}
				LogError("Loglevel changed to : ", *LogLevel)
			case syscall.SIGINT:
				gracefulShutdown()
			}
		}
	}
}

func gracefulShutdown() {
	log.Println("Graceful Shutdown")
	//RemoveFromConfig(svc)

	//would prefer to unregister HTTP and RPC handlers
	//need to figure out how to do that
	syscall.Sleep(10e9) // wait 10 seconds for requests to finish  #HACK
	syscall.Exit(0)
}

// Method to register the heartbeat of each skynet
// node with the healthcheck exporter.
func RegisterHeartbeat() {
	NewRpcServer("CommonService")
	// No AddToConfig()?
}


// A Generic struct to represent any process in the SkyNet system.
type Agent struct {
	Name    string
	Servers []*RpcServer
	chans   []chan bool
}

// Start the Agent and any servers.
// Block til all servers are done.
// This function is also responsible for
// registering the Heartbeat to healthcheck the service.
func (self *Agent) Start() *Agent {
	go watchSignals()
	go WatchConfig()
	RegisterHeartbeat()

	for _, server := range self.Servers {
		done := make(chan bool)
		self.chans = append(self.chans, done)
		go server.Serve(done)
	}
	return self
}

// Wait for all servers to finish.
func (self *Agent) Wait() {
	for _, done := range self.chans {
		<-done
	}
	self.chans = make([]chan bool, 0)
}

// Register the methods of the given sig type for a
// Skynet Server.
func (self *Agent) Register(sig interface{}) *Agent {
	server := NewRpcServer(sig)
	AddToConfig(server)
	self.Servers = append(self.Servers, server)
	return self
}

//Connect to the global config repo.
func NewAgent() *Agent {
	name := *Name
	initLogging()
	initDefaultExpVars(name)

	ConfigConnect()
	LoadConfig()
	if x := recover(); x != nil {
		LogWarn("No Configuration File loaded.  Creating One.")
	}

	node := &Agent{Name: name}

	return node
}
