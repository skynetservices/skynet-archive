package daemon

import (
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/client"
)

type Client struct {
	client.ServiceClientProvider
	requestInfo *skynet.RequestInfo
}

func GetDaemonForService(s *skynet.ServiceInfo) (c Client) {
	return GetDaemonForHost(s.ServiceAddr.IPAddress)
}

func GetDaemonForHost(host string) (c Client) {
	registered := true
	criteria := &skynet.Criteria{
		Hosts:      []string{host},
		Registered: &registered,
		Services: []skynet.ServiceCriteria{
			skynet.ServiceCriteria{Name: "SkynetDaemon"},
		},
	}

	s := client.GetServiceFromCriteria(criteria)
	c = Client{s, nil}
	return
}

func (c Client) ListSubServices(in ListSubServicesRequest) (out ListSubServicesResponse, err error) {
	err = c.Send(c.requestInfo, "ListSubServices", in, &out)
	return
}

func (c Client) StopAllSubServices(in StopAllSubServicesRequest) (out StopAllSubServicesResponse, err error) {
	err = c.Send(c.requestInfo, "StopAllSubServices", in, &out)
	return
}

func (c Client) StartSubService(in StartSubServiceRequest) (out StartSubServiceResponse, err error) {
	err = c.Send(c.requestInfo, "StartSubService", in, &out)
	return
}

func (c Client) StopSubService(in StopSubServiceRequest) (out StopSubServiceResponse, err error) {
	err = c.Send(c.requestInfo, "StopSubService", in, &out)
	return
}

func (c Client) RestartSubService(in RestartSubServiceRequest) (out RestartSubServiceResponse, err error) {
	err = c.Send(c.requestInfo, "RestartSubService", in, &out)
	return
}

func (c Client) RestartAllSubServices(in RestartAllSubServicesRequest) (out RestartAllSubServicesResponse, err error) {
	err = c.Send(c.requestInfo, "RestartAllSubServices", in, &out)
	return
}

func (c Client) RegisterSubService(in RegisterSubServiceRequest) (out RegisterSubServiceResponse, err error) {
	err = c.Send(c.requestInfo, "RegisterSubService", in, &out)
	return
}

func (c Client) UnregisterSubService(in UnregisterSubServiceRequest) (out UnregisterSubServiceResponse, err error) {
	err = c.Send(c.requestInfo, "UnregisterSubService", in, &out)
	return
}

func (c Client) SubServiceLogLevel(in SubServiceLogLevelRequest) (out SubServiceLogLevelResponse, err error) {
	err = c.Send(c.requestInfo, "SubServiceLogLevel", in, &out)
	return
}

func (c Client) LogLevel(in LogLevelRequest) (out LogLevelResponse, err error) {
	err = c.Send(c.requestInfo, "LogLevel", in, &out)
	return
}

func (c Client) Stop(in StopRequest) (out StopResponse, err error) {
	err = c.Send(c.requestInfo, "Stop", in, &out)
	c.Close()
	return
}
