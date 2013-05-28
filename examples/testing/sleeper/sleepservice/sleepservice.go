package main

import (
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/examples/testing/sleeper"
	"github.com/skynetservices/skynet/service"
	"log"
	"os"
	"time"
)

type Sleeper struct {
	service *service.Service
}

func NewSleeper() (f *Sleeper) {
	f = new(Sleeper)

	return
}

func (f *Sleeper) Registered(s *service.Service)                            {}
func (f *Sleeper) Unregistered(s *service.Service)                          {}
func (f *Sleeper) Started(s *service.Service)                               {}
func (f *Sleeper) Stopped(s *service.Service)                               {}
func (f *Sleeper) MethodCalled(method string)                               {}
func (f *Sleeper) MethodCompleted(method string, duration int64, err error) {}

func (f *Sleeper) Sleep(ri *skynet.RequestInfo, req sleeper.Request,
	resp *sleeper.Response) (err error) {

	log.Println("sleeping for", req.Duration, req.Message)

	if req.UnregisterHalfwayThrough {
		go func() {
			time.Sleep(req.Duration / 2)
			f.service.Unregister()
		}()
	}

	time.Sleep(req.Duration)

	resp.Message = req.Message

	if req.UnregisterWhenDone {
		f.service.Unregister()
	}

	if req.PanicWhenDone {
		panic("panic requested by client")
	}

	if req.ExitWhenDone {
		os.Exit(0)
	}

	return
}

func main() {
	f := NewSleeper()

	config, _ := skynet.GetServiceConfig()

	if config.Name == "" {
		config.Name = "Sleeper"
	}

	if config.Version == "unknown" {
		config.Version = "1"
	}

	if config.Region == "unknown" {
		config.Region = "Jersey"
	}

	clogger := skynet.NewConsoleSemanticLogger("sleepservice", os.Stdout)
	config.Log = skynet.NewMultiSemanticLogger(clogger)
	f.service = service.CreateService(f, config)

	// handle panic so that we remove ourselves from the pool in case
	// of catastrophic failure
	defer func() {
		f.service.Shutdown()
		if err := recover(); err != nil {
			log.Println("Unrecovered error occured: ", err)
		}
	}()

	// If we pass false here service will not be Registered we could
	// do other work/tasks by implementing the Started method and
	// calling Register() when we're ready
	waiter := f.service.Start(true)

	// waiting on the sync.WaitGroup returned by service.Start() will
	// wait for the service to finish running.
	waiter.Wait()
}
