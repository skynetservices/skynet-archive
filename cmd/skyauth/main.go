package main

import (
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/auth"
	"github.com/bketelsen/skynet/service"
	"log"
	"os"
)

type Authenticator struct {
	service *service.Service
}

func NewAuthenticator() (auth *Authenticator) {
	auth = new(Authenticator)

	return
}

func (auth *Authenticator) Registered(s *service.Service)   {}
func (auth *Authenticator) Unregistered(s *service.Service) {}
func (auth *Authenticator) Started(s *service.Service)      {}
func (auth *Authenticator) Stopped(s *service.Service)      {}

func (auth *Authenticator) Authenticate(ri *skynet.RequestInfo, req auth.AuthenticateRequest, resp *auth.AuthenticateResponse) (err error) {
	// TODO: actually query the db to authenticate, cache results, provide way to invalidate cache

	resp.Ok = true

	return
}

func main() {

	config, _ := skynet.GetServiceConfig()

	if config.Name == "" {
		config.Name = "Authenticator"
	}

	if config.Version == "unknown" {
		config.Version = "1"
	}

	if config.Region == "unknown" {
		config.Region = "Jersey"
	}

	var err error
	mlogger, err := skynet.NewMongoLogger("localhost", "skynet", "log", config.UUID)
	clogger := skynet.NewConsoleLogger("skyauth", os.Stdout)
	config.Log = skynet.NewMultiLogger(mlogger, clogger)
	if err != nil {
		config.Log.Item("Could not connect to mongo db for logging")
	}
	as := &Authenticator{}
	as.service = service.CreateService(as, config)

	// handle panic so that we remove ourselves from the pool in case of catastrophic failure
	defer func() {
		as.service.Shutdown()
		if err := recover(); err != nil {
			log.Println("Unrecovered error occured: ", err)
		}
	}()

	// If we pass false here service will not be Registered
	// we could do other work/tasks by implementing the Started method and calling Register() when we're ready
	waiter := as.service.Start(true)

	// waiting on the sync.WaitGroup returned by service.Start() will wait for the service to finish running.
	waiter.Wait()
}
