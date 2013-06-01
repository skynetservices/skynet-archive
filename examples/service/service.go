package main

import (
	"fmt"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/service2"
	"strings"
)

type TestService struct{}

func (s *TestService) Registered(service *service.Service)   {}
func (s *TestService) Unregistered(service *service.Service) {}
func (s *TestService) Started(service *service.Service)      {}
func (s *TestService) Stopped(service *service.Service) {
}

func NewTestService() *TestService {
	r := &TestService{}
	return r
}

func (s *TestService) Upcase(requestInfo *skynet.RequestInfo, in map[string]interface{}, out map[string]interface{}) (err error) {
	out["data"] = strings.ToUpper(in["data"].(string))
	return
}

func main() {
	testService := NewTestService()

	config, _ := skynet.GetServiceConfig()

	if config.Name == "" {
		config.Name = "TestService"
	}

	if config.Version == "" {
		config.Version = "1"
	}

	if config.Region == "unknown" {
		config.Region = "Clearwater"
	}

	service := service.CreateService(testService, config)

	// handle panic so that we remove ourselves from the pool in case
	// of catastrophic failure
	defer func() {
		service.Shutdown()
		if err := recover(); err != nil {
			fmt.Println("Unrecovered error occured: ", err)
		}
	}()

	// If we pass false here service will not be Registered we could
	// do other work/tasks by implementing the Started method and
	// calling Register() when we're ready
	fmt.Println("test")
	waiter := service.Start(true)
	fmt.Println("test2")

	// waiting on the sync.WaitGroup returned by service.Start() will
	// wait for the service to finish running.
	waiter.Wait()
}
