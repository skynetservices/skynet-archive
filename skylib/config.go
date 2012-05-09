//Copyright (c) 2011 Brian Ketelsen

//Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package skylib

import (
	"encoding/json"
	"expvar"
	"flag"
	"fmt"
	"github.com/ha/doozer"
	"log"
	"math/rand"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var DC *doozer.Conn
var NS *NetworkServers
var RpcServices []*RpcService

var Port *int = flag.Int("port", 9999, "tcp port to listen")
var Name *string = flag.String("name", "changeme", "name of this server")
var BindIP *string = flag.String("bindaddress", "127.0.0.1", "address to bind")
var LogFileName *string = flag.String("logFileName", "myservice.log", "name of logfile")
var LogLevel *int = flag.Int("logLevel", 1, "log level (1-5)")
var DoozerServer *string = flag.String("doozerServer", "127.0.0.1:8046", "addr:port of doozer server")
var Requests *expvar.Int
var Errors *expvar.Int
var Goroutines *expvar.Int
var svc *Service

// This is simple today - it returns the first listed service that matches the request
// Load balancing needs to be applied here somewhere.
func GetRandomClientByProvides(provides string) (*rpc.Client, error) {
	var providesList = make([]*Service, 0)

	var newClient *rpc.Client
	var err error

	for _, v := range NS.Services {
		if v != nil {
			if v.Provides == provides {
				providesList = append(providesList, v)
			}

		}
	}

	if len(providesList) > 0 {
		random := rand.Int() % len(providesList)
		s := providesList[random]

		portString := fmt.Sprintf("%s:%d", s.IPAddress, s.Port)
		newClient, err = rpc.DialHTTP("tcp", portString)
		if err != nil {
			LogWarn(fmt.Sprintf("Found %d Clients to service %s request.", len(providesList), provides))
			return nil, NewError(NO_CLIENT_PROVIDES_SERVICE, provides)
		}

	} else {
		return nil, NewError(NO_CLIENT_PROVIDES_SERVICE, provides)
	}
	return newClient, nil
}

func DoozerConnect() {
	var err error
	DC, err = doozer.Dial(*DoozerServer)
	if err != nil {
		log.Panic(err.Error())
	}
}

// on startup load the configuration file. 
// After the config file is loaded, we set the global config file variable to the
// unmarshaled data, making it useable for all other processes in this app.
func LoadConfig() {
	rev, err := DC.Rev()
	if err != nil {
		log.Panic(err.Error())
	}
	names, err := DC.Getdir("/services/", rev, 0, -1)

	for _, name := range names {
		data, _, err := DC.Get("/services/"+name, nil)
		if err != nil {
			log.Panic(err.Error())
		}
		if len(data) > 0 {
			setConfig(data)
			return
		}

	}
}

func (r *Service) RemoveFromConfig() {

	rev, err := DC.Rev()
	if err != nil {
		log.Panic(err.Error())
	}
	err = DC.Del("/services/"+r.Name, rev)
	if err != nil {
		log.Panic(err.Error())
	}
}

func (r *Service) AddToConfig() {
	/*	for _, v := range NS.Services {
			if v != nil {
				if v.Equal(r) {
					LogInfo(fmt.Sprintf("Skipping adding %s : already exists.", v.Name))
					return // it's there so we don't need an update
				}
			}
		}
		NS.Services = append(NS.Services, r)
	*/
	b, err := json.Marshal(r)
	if err != nil {
		log.Panic(err.Error())
	}
	rev, err := DC.Rev()
	if err != nil {
		log.Panic(err.Error())
	}
	_, err = DC.Set("/services/"+r.Name, rev, b)
	if err != nil {
		log.Panic(err.Error())
	}
}

// unmarshal data from remote store into global config variable
func setConfig(data []byte) {
	var svc Service
	err := json.Unmarshal(data, &svc)
	fmt.Println(svc)
	if err != nil {
		log.Panic(err.Error())
	}
}

// Watch for remote changes to the config file.  When new changes occur
// reload our copy of the config file.
// Meant to be run as a goroutine continuously.
func WatchConfig() {
	rev, err := DC.Rev()
	if err != nil {
		log.Panic(err.Error())
	}
	for {

		// blocking wait call returns on a change
		ev, err := DC.Wait("/servers/*", rev)
		if err != nil {
			log.Panic(err.Error())
		}
		log.Println("Received new configuration.  Setting local config.")
		setConfig(ev.Body)

		rev = ev.Rev + 1
	}

}

func initDefaultExpVars(name string) {
	Requests = expvar.NewInt(name + "-processed")
	Errors = expvar.NewInt(name + "-errors")
	Goroutines = expvar.NewInt(name + "-goroutines")
}

func watchSignals(c chan os.Signal) {

	for {
		select {
		case sig := <-c:
			switch sig.(syscall.Signal) {
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
	svc.RemoveFromConfig()

	//would prefer to unregister HTTP and RPC handlers
	//need to figure out how to do that
	time.Sleep(10e9) // wait 10 seconds for requests to finish  #HACK
	syscall.Exit(0)
}

func Setup(name string) {
	DoozerConnect()
	LoadConfig()
	if x := recover(); x != nil {
		LogWarn("No Configuration File loaded.  Creating One.")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGUSR1, syscall.SIGUSR2)
	go watchSignals(c)

	initDefaultExpVars(name)

	svc = NewService(name)

	svc.AddToConfig()

	go WatchConfig()

	RegisterHeartbeat()

}
