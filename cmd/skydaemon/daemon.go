package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/config"
	"github.com/skynetservices/skynet/daemon"
	"github.com/skynetservices/skynet/log"
	"github.com/skynetservices/skynet/service"
	"github.com/skynetservices/skynet/stats"
	"io/ioutil"
	"os"
	"runtime"
	"sync"
	"time"
)

// SkynetDaemon is a service for administering other services
type SkynetDaemon struct {
	Services    map[string]*SubService
	serviceLock sync.Mutex       `json:"-"`
	Service     *service.Service `json:"-"`
	HostStats   stats.Host       `json:"-"`

	stateFile *os.File `json:"-"`
	saveChan  chan bool
}

func NewSkynetDaemon() *SkynetDaemon {
	sFile := stateFileName()

	if _, err := os.Stat(sFile); os.IsNotExist(err) {
		panic("state file is missing:" + sFile)
	}

	f, err := os.OpenFile(sFile, os.O_RDWR|os.O_CREATE, 0660)

	if err != nil {
		panic("could not open state file" + sFile)
	}

	d := &SkynetDaemon{
		Services:  map[string]*SubService{},
		stateFile: f,
		saveChan:  make(chan bool, 1),
	}

	go d.mux()

	return d
}

func (sd *SkynetDaemon) Registered(s *service.Service)   {}
func (sd *SkynetDaemon) Unregistered(s *service.Service) {}
func (sd *SkynetDaemon) Started(s *service.Service) {
	err := sd.cleanupHost(s.ServiceInfo.UUID)
	if err != nil {
		log.Println(log.ERROR, "Error cleaning up host", err)
	}

	err = sd.restoreState()

	if err != nil {
		log.Println(log.ERROR, "Error restoring state", err)
	}
}

func (sd *SkynetDaemon) Stopped(s *service.Service) {
}

func (s *SkynetDaemon) StartSubService(requestInfo *skynet.RequestInfo, in daemon.StartSubServiceRequest, out *daemon.StartSubServiceResponse) (err error) {
	out.UUID = config.NewUUID()

	log.Printf(log.TRACE, "%+v", SubserviceStart{
		BinaryName: in.BinaryName,
		Args:       in.Args,
	})

	ss, err := NewSubService(in.BinaryName, in.Args, out.UUID, in.Registered)
	if err != nil {
		return
	}

	s.serviceLock.Lock()
	s.Services[out.UUID] = ss
	s.serviceLock.Unlock()

	start, startErr := ss.Start()

	if startErr != nil {
		return errors.New("Service failed to start: " + startErr.Error())
	} else if !start {
		return errors.New("Service failed to start")
	}

	tc := time.Tick(RerunWait * 2)

	go func() {
		// Wait for startup timer to see if we're still running
		// We want to avoid keeping a state of a large list of services that failed to start
		<-tc

		if ss.IsRunning() {
			s.saveState()
		}
	}()

	return
}

func (s *SkynetDaemon) updateHostStats(host string) {
	s.HostStats.Update(host)
}

func (s *SkynetDaemon) getSubService(uuid string) (ss *SubService) {
	s.serviceLock.Lock()
	ss = s.Services[uuid]
	s.serviceLock.Unlock()
	return
}

func (s *SkynetDaemon) ListSubServices(requestInfo *skynet.RequestInfo, in daemon.ListSubServicesRequest, out *daemon.ListSubServicesResponse) (err error) {
	out.Services = make(map[string]daemon.SubServiceInfo)
	if len(s.Services) == 0 {
		err = errors.New("No services started")
		return
	}
	for uuid, ss := range s.Services {
		out.Services[uuid] = daemon.SubServiceInfo{
			UUID:        uuid,
			ServicePath: ss.ServicePath,
			Args:        ss.Args,
			Running:     ss.running,
		}
	}
	fmt.Println(out)
	return
}

func (s *SkynetDaemon) StopAllSubServices(requestInfo *skynet.RequestInfo, in daemon.StopAllSubServicesRequest, out *daemon.StopAllSubServicesResponse) (err error) {
	var uuids []string
	s.serviceLock.Lock()
	for uuid := range s.Services {
		uuids = append(uuids, uuid)
	}
	s.serviceLock.Unlock()

	out.Stops = make([]daemon.StopSubServiceResponse, len(uuids))

	for i, uuid := range uuids {
		log.Println(log.TRACE, "Stopping "+uuid)
		err = s.StopSubService(requestInfo, daemon.StopSubServiceRequest{UUID: uuid}, &out.Stops[i])
		if err != nil {
			log.Println(log.ERROR, "Failed to stop subservice "+uuid, err)
			return
		}
		if out.Stops[i].Ok {
			out.Count++
		}
	}

	s.saveState()

	return
}

func (s *SkynetDaemon) StopSubService(requestInfo *skynet.RequestInfo, in daemon.StopSubServiceRequest, out *daemon.StopSubServiceResponse) (err error) {
	ss := s.getSubService(in.UUID)
	if ss != nil {
		out.Ok = ss.Stop()
		out.UUID = in.UUID
		delete(s.Services, in.UUID)

		s.saveState()
	} else {
		err = errors.New(fmt.Sprintf("No such service UUID %q", in.UUID))
	}

	return
}

func (s *SkynetDaemon) RegisterSubService(requestInfo *skynet.RequestInfo, in daemon.RegisterSubServiceRequest, out *daemon.RegisterSubServiceResponse) (err error) {
	ss := s.getSubService(in.UUID)
	if ss != nil {
		out.Ok = ss.Register()
		out.UUID = in.UUID

		s.saveState()
	} else {
		err = errors.New(fmt.Sprintf("No such service UUID %q", in.UUID))
	}

	return
}

func (s *SkynetDaemon) UnregisterSubService(requestInfo *skynet.RequestInfo, in daemon.UnregisterSubServiceRequest, out *daemon.UnregisterSubServiceResponse) (err error) {
	ss := s.getSubService(in.UUID)
	if ss != nil {
		out.Ok = ss.Unregister()
		out.UUID = in.UUID

		s.saveState()
	} else {
		err = errors.New(fmt.Sprintf("No such service UUID %q", in.UUID))
	}

	return
}

func (s *SkynetDaemon) RestartSubService(requestInfo *skynet.RequestInfo, in daemon.RestartSubServiceRequest, out *daemon.RestartSubServiceResponse) (err error) {
	ss := s.getSubService(in.UUID)
	if ss != nil {
		ss.Restart()
		out.UUID = in.UUID

		s.saveState()
	} else {
		err = errors.New(fmt.Sprintf("No such service UUID %q", in.UUID))
	}
	return
}

func (s *SkynetDaemon) RestartAllSubServices(requestInfo *skynet.RequestInfo, in daemon.RestartAllSubServicesRequest, out *daemon.RestartAllSubServicesResponse) (err error) {
	var uuids []string
	s.serviceLock.Lock()
	for uuid := range s.Services {
		uuids = append(uuids, uuid)
	}
	s.serviceLock.Unlock()

	if len(uuids) == 0 {
		err = errors.New("No services started")
		return
	}

	out.Restarts = make([]daemon.RestartSubServiceResponse, len(uuids))

	for i, uuid := range uuids {
		err = s.RestartSubService(requestInfo, daemon.RestartSubServiceRequest{UUID: uuid}, &out.Restarts[i])
		if err != nil {
			return
		}
	}

	s.saveState()
	return
}

func (s *SkynetDaemon) SubServiceLogLevel(requestInfo *skynet.RequestInfo, in daemon.SubServiceLogLevelRequest, out *daemon.SubServiceLogLevelResponse) (err error) {
	ss := s.getSubService(in.UUID)
	if ss != nil {
		out.Ok = ss.SetLogLevel(in.Level)
		out.UUID = in.UUID
		out.Level = in.Level
	} else {
		err = errors.New(fmt.Sprintf("No such service UUID %q", in.UUID))
	}

	return
}

func (s *SkynetDaemon) LogLevel(requestInfo *skynet.RequestInfo, in daemon.LogLevelRequest, out *daemon.LogLevelResponse) (err error) {
	log.SetLogLevel(log.LevelFromString(in.Level))
	out.Ok = true
	out.Level = in.Level

	return
}

func (s *SkynetDaemon) Stop(requestInfo *skynet.RequestInfo, in daemon.StopRequest, out *daemon.StopResponse) (err error) {
	out.Ok = true

	s.serviceLock.Lock()
	for _, ss := range s.Services {
		ss.Stop()
	}
	s.serviceLock.Unlock()
	go s.Service.Shutdown()

	s.saveState()

	return
}

func (s *SkynetDaemon) saveState() {
	select {
	case s.saveChan <- true:
		// Throw away save, there is one already queued
	default:
	}
}

func stateFileName() string {
	if os.Getenv("SKYNET_STATEFILE") != "" {
		return os.Getenv("SKYNET_STATEFILE")
	}

	if runtime.GOOS == "darwin" {
		return "/usr/local/lib/skynet/.skystate"
	}

	return "/var/lib/skynet/.skystate"
}

// TODO: This should be moved out so that it's run asynchronously
// it should also use a buffered channel so that if a save is already queued it only saves once
func (s *SkynetDaemon) writeStateFile() (err error) {
	err = s.stateFile.Truncate(0)

	if err != nil {
		return
	}

	_, err = s.stateFile.Seek(0, 0)

	if err != nil {
		return
	}

	var b []byte
	b, err = json.MarshalIndent(s.Services, "", "\t")

	if err != nil {
		log.Println(log.ERROR, "Failed to marshall daemon state")
		return
	}

	_, err = s.stateFile.Write(b)

	if err != nil {
		log.Println(log.ERROR, "Failed to save daemon state")
	}

	return
}

func (s *SkynetDaemon) restoreState() (err error) {
	var b []byte
	b, err = ioutil.ReadAll(s.stateFile)

	if err != nil {
		return
	}

	// no state to restore
	if len(b) == 0 {
		return
	}

	services := make(map[string]*SubService)
	err = json.Unmarshal(b, &services)

	if err != nil {
		return
	}

	for _, service := range services {
		var ss *SubService
		ss, err = NewSubService(service.ServicePath, service.Args, service.UUID, service.Registered)
		if err != nil {
			return
		}

		s.serviceLock.Lock()
		s.Services[service.UUID] = ss
		s.serviceLock.Unlock()

		start, startErr := ss.Start()

		if startErr != nil {
			return errors.New("Service failed to start: " + startErr.Error())
		} else if !start {
			return errors.New("Service failed to start")
		}
	}

	return
}

func (s *SkynetDaemon) closeStateFile() {
	if s.stateFile != nil {
		s.stateFile.Close()
	}
}

func (s *SkynetDaemon) mux() {
	for {
		select {
		case <-s.saveChan:
			s.writeStateFile()
		}
	}
}

func (s *SkynetDaemon) cleanupHost(daemonUUID string) (err error) {
	sm := skynet.GetServiceManager()
	c := skynet.Criteria{}

	c.AddHost(s.Service.ServiceInfo.ServiceAddr.IPAddress)

	var instances []skynet.ServiceInfo
	instances, err = sm.ListInstances(&c)

	if err != nil {
		return
	}

	for _, i := range instances {
		if i.UUID != daemonUUID {
			sm.Remove(i)
		}
	}

	return
}
