package daemon

import (
	"github.com/skynetservices/skynet2"
	"github.com/skynetservices/skynet2/client"
)

type Client struct {
	*client.ServiceClient
	requestInfo *skynet.RequestInfo
}

func GetDaemonForService(cl *client.Client, s *skynet.ServiceInfo) (c Client) {
	return GetDaemonForHost(cl, s.ServiceAddr.IPAddress)
}

func GetDaemonForHost(cl *client.Client, host string) (c Client) {
	registered := true
	criteria := &skynet.Criteria{
		Hosts:      []string{host},
		Registered: &registered,
		Services: []skynet.ServiceCriteria{
			skynet.ServiceCriteria{Name: "SkynetDaemon"},
		},
	}

	s := cl.GetService(criteria)
	c = Client{s, nil}
	return
}

func (c Client) Start(in StartRequest) (out StartResponse, err error) {
	err = c.Send(c.requestInfo, "Start", in, &out)
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

func (c Client) StartAllSubServices(in StartAllSubServicesRequest) (out StartAllSubServicesResponse, err error) {
	err = c.Send(c.requestInfo, "StartAllSubServices", in, &out)
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
