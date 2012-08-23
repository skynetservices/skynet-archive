package daemon

import (
	"errors"
	"fmt"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/service"
	"sync"
)

// SkynetDaemon is a service for administering other services
type SkynetDaemon struct {
	Log         skynet.Logger
	Services    map[string]*SubService
	serviceLock sync.Mutex
}

func (sd *SkynetDaemon) Registered(s *service.Service)   {}
func (sd *SkynetDaemon) Unregistered(s *service.Service) {}
func (sd *SkynetDaemon) Started(s *service.Service)      {}
func (sd *SkynetDaemon) Stopped(s *service.Service) {
	sd.StopAllSubServices(&skynet.RequestInfo{}, StopAllSubServicesRequest{}, &StopAllSubServicesResponse{})
}

func (s *SkynetDaemon) Deploy(requestInfo *skynet.RequestInfo, in DeployRequest, out *DeployResponse) (err error) {
	out.UUID = skynet.UUID()

	s.Log.Item(SubserviceDeployment{
		ServicePath: in.ServicePath,
		Args:        in.Args,
	})

	ss, err := NewSubService(s.Log, in.ServicePath, in.Args, out.UUID)
	if err != nil {
		return
	}
	s.serviceLock.Lock()
	s.Services[out.UUID] = ss
	s.serviceLock.Unlock()
	return
}

func (s *SkynetDaemon) getSubService(uuid string) (ss *SubService) {
	s.serviceLock.Lock()
	ss = s.Services[uuid]
	s.serviceLock.Unlock()
	return
}

type M map[string]interface{}

func (m M) String(key string) (val string, ok bool) {
	vali, ok := m[key]
	if !ok {
		return
	}
	val, ok = vali.(string)
	return
}

func (s *SkynetDaemon) ListSubServices(requestInfo *skynet.RequestInfo, in ListSubServicesRequest, out *ListSubServicesResponse) (err error) {
	out.Services = make(map[string]SubServiceInfo)
	if len(s.Services) == 0 {
		err = errors.New("No services deployed")
		return
	}
	for uuid, ss := range s.Services {
		out.Services[uuid] = SubServiceInfo{
			UUID:        uuid,
			ServicePath: ss.ServicePath,
			Args:        ss.Args,
			Running:     ss.running,
		}
	}
	fmt.Println(out)
	return
}

func (s *SkynetDaemon) StopAllSubServices(requestInfo *skynet.RequestInfo, in StopAllSubServicesRequest, out *StopAllSubServicesResponse) (err error) {
	var uuids []string
	s.serviceLock.Lock()
	for uuid := range s.Services {
		uuids = append(uuids, uuid)
	}
	s.serviceLock.Unlock()

	out.Stops = make([]StopSubServiceResponse, len(uuids))

	for i, uuid := range uuids {
		err = s.StopSubService(requestInfo, StopSubServiceRequest{UUID: uuid}, &out.Stops[i])
		if err != nil {
			return
		}
		if out.Stops[i].Ok {
			out.Count++
		}
	}
	return
}

func (s *SkynetDaemon) StartAllSubServices(requestInfo *skynet.RequestInfo, in StartAllSubServicesRequest, out *StartAllSubServicesResponse) (err error) {
	var uuids []string
	s.serviceLock.Lock()
	for uuid := range s.Services {
		uuids = append(uuids, uuid)
	}
	s.serviceLock.Unlock()

	if len(uuids) == 0 {
		err = errors.New("No services deployed")
		return
	}

	out.Starts = make([]StartSubServiceResponse, len(uuids))

	for i, uuid := range uuids {
		err = s.StartSubService(requestInfo, StartSubServiceRequest{UUID: uuid}, &out.Starts[i])
		if err != nil {
			return
		}
		if out.Starts[i].Ok {
			out.Count++
		}
	}
	return
}

func (s *SkynetDaemon) StartSubService(requestInfo *skynet.RequestInfo, in StartSubServiceRequest, out *StartSubServiceResponse) (err error) {
	ss := s.getSubService(in.UUID)
	if ss != nil {
		out.Ok = ss.Start()
		out.UUID = in.UUID
	} else {
		err = errors.New(fmt.Sprintf("No such service UUID %q", in.UUID))
	}
	return
}

func (s *SkynetDaemon) StopSubService(requestInfo *skynet.RequestInfo, in StopSubServiceRequest, out *StopSubServiceResponse) (err error) {
	ss := s.getSubService(in.UUID)
	out.Ok = ss.Stop()
	out.UUID = in.UUID
	return
}

func (s *SkynetDaemon) RestartSubService(requestInfo *skynet.RequestInfo, in RestartSubServiceRequest, out *RestartSubServiceResponse) (err error) {
	ss := s.getSubService(in.UUID)
	ss.Restart()
	out.UUID = in.UUID
	return
}

func (s *SkynetDaemon) RestartAllSubServices(requestInfo *skynet.RequestInfo, in RestartAllSubServicesRequest, out *RestartAllSubServicesResponse) (err error) {
	var uuids []string
	s.serviceLock.Lock()
	for uuid := range s.Services {
		uuids = append(uuids, uuid)
	}
	s.serviceLock.Unlock()

	if len(uuids) == 0 {
		err = errors.New("No services deployed")
		return
	}

	out.Restarts = make([]RestartSubServiceResponse, len(uuids))

	for i, uuid := range uuids {
		err = s.RestartSubService(requestInfo, RestartSubServiceRequest{UUID: uuid}, &out.Restarts[i])
		if err != nil {
			return
		}
	}
	return
}
