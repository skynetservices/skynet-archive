package skylib

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"syscall"
)

// TODO: Better error handling, should gracefully fail to startup if it can't connect to doozer

// A Generic struct to represent any service in the SkyNet system.
type ServiceInterface interface {
	Started(s *Service)
	Stopped(s *Service)
	Registered(s *Service)
	Unregistered(s *Service)
}

type Service struct {
	Config     *ServiceConfig
	DoozerConn DoozerConnection `json:"-"`
	Registered bool
	doneChan   chan bool `json:"-"`

	Log Logger `json:"-"`

	Delegate ServiceInterface         `json:"-"`
	methods  map[string]reflect.Value `json:"-"`
}

func (s *Service) Resolve(name string, arguments []reflect.Value) (interface{}, reflect.Value, error) {
	return s.Delegate, s.methods[name], nil
}

func (s *Service) Start(register bool) {
	portString := fmt.Sprintf("%s:%d", s.Config.ServiceAddr.IPAddress, s.Config.ServiceAddr.Port)

	rpcServ := NewRpcServer(s, true, nil)
	l, _ := net.Listen("tcp", portString)

	rpcServ.Listen(l)

	// Watch signals for shutdown
	c := make(chan os.Signal, 1)
	go watchSignals(c, s)

	s.doneChan = make(chan bool, 1)
	s.Log.Item("Starting server")

	go rpcServ.Run()

	go s.Delegate.Started(s) // Call user defined callback

	s.UpdateCluster()

	if register == true {
		s.Register()
	}

	// Endless loop to keep app from returning
	select {
	case _ = <-s.doneChan:
		//NOTE: probably shouldn't call Exit() in a lib. Just let the function return?
		//      But then, this is triggered by a kill signal, so it's more like we 
		//      intercept the kill signal, clean up, and then die anyway.
		syscall.Exit(0)
	}
}

func (s *Service) UpdateCluster() {
	b, err := json.Marshal(s)
	if err != nil {
		s.Log.Panic(err.Error())
	}

	rev := s.doozer().GetCurrentRevision()

	_, err = s.doozer().Set(s.GetConfigPath(), rev, b)
	if err != nil {
		s.Log.Panic(err.Error())
	}
}

func (s *Service) RemoveFromCluster() {
	rev := s.doozer().GetCurrentRevision()
	path := s.GetConfigPath()
	err := s.doozer().Del(path, rev)
	if err != nil {
		s.Log.Panic(err.Error())
	}
}

func (s *Service) Register() {
	s.Registered = true
	s.UpdateCluster()

	s.Delegate.Registered(s) // Call user defined callback
}

func (s *Service) Unregister() {
	s.Registered = false
	s.UpdateCluster()

	s.Delegate.Unregistered(s) // Call user defined callback
}

func (s *Service) Shutdown() {
	s.Unregister()

	// TODO: make this wait for requests to finish
	s.RemoveFromCluster()
	s.doneChan <- true

	s.Delegate.Stopped(s) // Call user defined callback
}

func CreateService(s ServiceInterface, c *ServiceConfig) *Service {
	typ := reflect.TypeOf(s)

	// This will set defaults
	initializeConfig(c)

	service := &Service{
		Config:   c,
		Delegate: s,
		Log:      c.Log,
		methods:  make(map[string]reflect.Value),
	}

	c.Log.Item(c)

	service.findRPCMethods(typ)

	return service
}

func (s *Service) findRPCMethods(typ reflect.Type) {
	for i := 0; i < typ.NumMethod(); i++ {
		m := typ.Method(i)

		// Don't register callbacks
		if m.Name == "Started" || m.Name == "Stopped" || m.Name == "Registered" || m.Name == "Unregistered" {
			continue
		}

		// Only register exported methods
		if m.PkgPath != "" {
			continue
		}

		if m.Type.NumOut() != 1 && m.Type.NumOut() != 2 {
			continue
		}

		s.methods[m.Name] = m.Func
		//s.Log.Println("Registered RPC Method: " + m.Name)
		s.Log.ServiceItem(s, RegisteredMethod{
			Method: m.Name,
		})
	}
}

func initializeConfig(c *ServiceConfig) {
	if c.Log == nil {
		c.Log = NewConsoleLogger(os.Stderr)
	}

	if c.Name == "" {
		c.Name = "SkynetService"
	}

	if c.Version == "" {
		c.Version = "1"
	}

	if c.Region == "" {
		c.Region = "local"
	}

	if c.ServiceAddr.IPAddress == "" {
		c.ServiceAddr.IPAddress = "127.0.0.1"
	}

	if c.ServiceAddr.Port == 0 {
		c.ServiceAddr.Port = 9000
	}

	if c.DoozerConfig == nil {
		c.DoozerConfig = &DoozerConfig{
			Uri:          "127.0.0.1:8046",
			AutoDiscover: true,
		}
	}
}

func (s *Service) GetConfigPath() string {
	return "/services/" + s.Config.Name + "/" + s.Config.Version + "/" + s.Config.Region + "/" + s.Config.ServiceAddr.IPAddress + "/" + strconv.Itoa(s.Config.ServiceAddr.Port)
}

func (r *Service) Equal(that *Service) bool {
	var b bool
	b = false
	if r.Config.Name != that.Config.Name {
		return b
	}
	if r.Config.ServiceAddr.IPAddress != that.Config.ServiceAddr.IPAddress {
		return b
	}
	if r.Config.ServiceAddr.Port != that.Config.ServiceAddr.Port {
		return b
	}
	b = true
	return b
}

func watchSignals(c chan os.Signal, s *Service) {
	signal.Notify(c, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGSEGV, syscall.SIGSTOP, syscall.SIGTERM)

	for {
		select {
		case sig := <-c:
			switch sig.(syscall.Signal) {
			// Trap signals for clean shutdown
			case syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGSEGV, syscall.SIGSTOP, syscall.SIGTERM:
				s.Shutdown()
			}
		}
	}
}

func (s *Service) doozer() DoozerConnection {
	if s.DoozerConn == nil {
		s.DoozerConn = NewDoozerConnectionFromConfig(*s.Config.DoozerConfig, s.Log)

		s.DoozerConn.Connect()
	}

	return s.DoozerConn
}
