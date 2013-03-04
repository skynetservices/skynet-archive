package main

import (
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/service"
)

type TutorialService struct {
}

func (s *TutorialService) Registered(service *service.Service)                      {}
func (s *TutorialService) Unregistered(service *service.Service)                    {}
func (s *TutorialService) Started(service *service.Service)                         {}
func (s *TutorialService) Stopped(service *service.Service)                         {}
func (s *TutorialService) MethodCalled(method string)                               {}
func (s *TutorialService) MethodCompleted(method string, duration int64, err error) {}

type TutorialRequest struct {
	Value int
}

type TutorialResponse struct {
	Value int
}

func (f *TutorialService) AddOne(ri *skynet.RequestInfo, req *TutorialRequest, resp *TutorialResponse) (err error) {
	resp.Value = req.Value + 1

	return nil
}

func main() {
	tutorial := &TutorialService{}
	config, _ := skynet.GetServiceConfig()

	config.Name = "TutorialService"
	config.Version = "1"
	config.Region = "Development"

	service := service.CreateService(tutorial, config)

	defer func() {
		service.Shutdown()
	}()

	waiter := service.Start(true)
	waiter.Wait()
}
