package skylib

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"syscall"
)

// A Generic struct to represent any service in the SkyNet system.
type ServiceInterface interface {
	Started(s *Service)
	Stopped(s *Service)
	Registered(s *Service)
	UnRegistered(s *Service)
}

type Service struct {
	Config     *ServiceConfig
	DoozerConn *DoozerConnection `json:"-"`
	Registered bool              `json:"-"`
	doneChan   chan bool         `json:"-"`

	Log *log.Logger `json:"-"`

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
	s.Log.Println("Starting server")

	go rpcServ.Run()

	go s.Delegate.Started(s) // Call user defined callback

	if register == true {
		s.Register()
	}

	// Endless loop to keep app from returning
	select {
	case _ = <-s.doneChan:
	}
}

func (s *Service) Register() {

	// TODO: we need a different object to represent this, we don't need all these additional params being forwarded along
	b, err := json.Marshal(s)
	if err != nil {
		s.Log.Panic(err.Error())
	}

	rev := s.doozer().GetCurrentRevision()

	_, err = s.doozer().Set(s.GetConfigPath(), rev, b)
	if err != nil {
		s.Log.Panic(err.Error())
	}

	s.Registered = true

	s.Delegate.Registered(s) // Call user defined callback
}

func (s *Service) UnRegister() {
	if s.Registered == true {
		rev := s.doozer().GetCurrentRevision()
		path := s.GetConfigPath()
		err := s.doozer().Del(path, rev)
		if err != nil {
			s.Log.Panic(err.Error())
		}
	}

	s.Delegate.UnRegistered(s) // Call user defined callback
}

func (s *Service) Shutdown() {
	// TODO: make this wait for requests to finish
	s.UnRegister()
	s.doneChan <- true
	syscall.Exit(0)

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

	service.findRPCMethods(typ)

	return service
}

func (s *Service) findRPCMethods(typ reflect.Type) {
	for i := 0; i < typ.NumMethod(); i++ {
		m := typ.Method(i)

		// Don't register callbacks
		if m.Name == "Started" || m.Name == "Stopped" || m.Name == "Registered" || m.Name == "UnRegistered" {
			continue
		}

		// Only register exported methods
		if m.PkgPath != "" {
			continue
		}

		// TODO: Ensure method matches required signature
		if m.Type.NumOut() != 1 && m.Type.NumOut() != 2 {
			continue
		}

		s.methods[m.Name] = m.Func
		s.Log.Println("Registered RPC Method: " + m.Name)
	}
}

func initializeConfig(c *ServiceConfig) {
	if c.Log == nil {
		c.Log = log.New(os.Stderr, "", log.LstdFlags)
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
    c.DoozerConfig = &DoozerConfig {
      Uri: "127.0.0.1:8046",
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

func (s *Service) doozer() *DoozerConnection {
	if s.DoozerConn == nil {
		s.DoozerConn = &DoozerConnection {
			Config:  s.Config.DoozerConfig,
      Log: s.Log,
		}

		s.DoozerConn.Connect()
	}

	return s.DoozerConn
}
