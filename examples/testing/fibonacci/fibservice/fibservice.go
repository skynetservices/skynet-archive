package main

import (
	"errors"
	"fmt"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/client"
	"github.com/bketelsen/skynet/examples/testing/fibonacci"
	"github.com/bketelsen/skynet/service"
	"log"
	"os"
	"sync"
)

type Fibonacci struct {
	cconfig *skynet.ClientConfig
	client  *client.Client

	// previously computed values
	cache  map[int]chan uint64
	cmutex sync.Mutex
}

func NewFibonacci() (f *Fibonacci) {
	f = new(Fibonacci)

	f.cconfig, _ = skynet.GetClientConfig()
	f.client = client.NewClient(f.cconfig)

	f.cache = map[int]chan uint64{
		0: make(chan uint64, 1),
		1: make(chan uint64, 1),
	}
	f.cache[0] <- 0
	f.cache[1] <- 1

	return
}

func (f *Fibonacci) Registered(s *service.Service)   {}
func (f *Fibonacci) Unregistered(s *service.Service) {}
func (f *Fibonacci) Started(s *service.Service)      {}
func (f *Fibonacci) Stopped(s *service.Service)      {}

func (f *Fibonacci) Index(ri *skynet.RequestInfo, req fibonacci.Request,
	resp *fibonacci.Response) (err error) {

	if req.Index < 0 {
		err = errors.New(fmt.Sprintf("Invalid request: %+v", req))
		return
	}

	resp.Index = req.Index

	f.cmutex.Lock()
	vchan, ok := f.cache[req.Index]
	if ok {
		f.cmutex.Unlock()
		resp.Value = <-vchan
		vchan <- resp.Value
		return
	}
	f.cache[req.Index] = make(chan uint64, 1)
	f.cmutex.Unlock()

	v1ch := make(chan uint64)
	go f.lookupValue(ri, req.Index-1, v1ch)
	v2ch := make(chan uint64)
	go f.lookupValue(ri, req.Index-2, v2ch)

	resp.Value = <-v1ch
	resp.Value += <-v2ch

	f.cmutex.Lock()
	f.cache[req.Index] <- resp.Value
	f.cmutex.Unlock()

	return
}

func (f *Fibonacci) lookupValue(ri *skynet.RequestInfo, index int,
	vchan chan<- uint64) {

	remoteService := f.client.GetService("Fibonacci", "", "", "")

	var err error
	for {
		req := fibonacci.Request{
			Index: index,
		}
		resp := fibonacci.Response{}
		err = remoteService.Send(ri, "Index", req, &resp)

		if err == nil {
			vchan <- resp.Value
			return
		}
		f.client.Log.Error(err.Error())
	}
}

func main() {
	f := NewFibonacci()

	config, _ := skynet.GetServiceConfig()

	if config.Name == "" {
		config.Name = "Fibonacci"
	}

	if config.Version == "unknown" {
		config.Version = "1"
	}

	if config.Region == "unknown" {
		config.Region = "Jersey"
	}

	var err error
	mlogger, err := skynet.NewMongoSemanticLogger("localhost", "skynet",
		"log", config.UUID, config)
	clogger := skynet.NewConsoleSemanticLogger("fibservice", os.Stdout)
	config.Log = skynet.NewMultiSemanticLogger(mlogger, clogger)
	if err != nil {
		config.Log.Error("Could not connect to mongo db for logging")
	}
	service := service.CreateService(f, config)

	// handle panic so that we remove ourselves from the pool in case
	// of catastrophic failure
	defer func() {
		service.Shutdown()
		if err := recover(); err != nil {
			log.Println("Unrecovered error occured: ", err)
		}
	}()

	// If we pass false here service will not be Registered we could
	// do other work/tasks by implementing the Started method and
	// calling Register() when we're ready
	waiter := service.Start(true)

	// waiting on the sync.WaitGroup returned by service.Start() will
	// wait for the service to finish running.
	waiter.Wait()
}
