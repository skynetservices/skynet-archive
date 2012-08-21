package main

import (
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/examples/testing/sleeper"
	"github.com/bketelsen/skynet/service"
	"log"
	"os"
	"time"
)

type Sleeper struct {
}

func NewSleeper() (f *Sleeper) {
	f = new(Sleeper)

	return
}

func (f *Sleeper) Registered(s *service.Service)   {}
func (f *Sleeper) Unregistered(s *service.Service) {}
func (f *Sleeper) Started(s *service.Service)      {}
func (f *Sleeper) Stopped(s *service.Service)      {}

func (f *Sleeper) Sleep(ri *skynet.RequestInfo, req sleeper.Request, resp *sleeper.Response) (err error) {
	time.Sleep(1 * time.Second)

	resp.Message = req.Message

	return
}

func main() {
	f := NewSleeper()

	config, _ := skynet.GetServiceConfigFromFlags()

	if config.Name == "" {
		config.Name = "Sleeper"
	}

	if config.Version == "unknown" {
		config.Version = "1"
	}

	if config.Region == "unknown" {
		config.Region = "Jersey"
	}

	var err error
	mlogger, err := skynet.NewMongoLogger("localhost", "skynet", "log", config.UUID)
	clogger := skynet.NewConsoleLogger(os.Stdout)
	config.Log = skynet.NewMultiLogger(mlogger, clogger)
	if err != nil {
		config.Log.Item("Could not connect to mongo db for logging")
	}
	service := service.CreateService(f, config)

	// handle panic so that we remove ourselves from the pool in case of catastrophic failure
	defer func() {
		service.Shutdown()
		if err := recover(); err != nil {
			log.Println("Unrecovered error occured: ", err)
		}
	}()

	// If we pass false here service will not be Registered
	// we could do other work/tasks by implementing the Started method and calling Register() when we're ready
	waiter := service.Start(true)

	// waiting on the sync.WaitGroup returned by service.Start() will wait for the service to finish running.
	waiter.Wait()
}
