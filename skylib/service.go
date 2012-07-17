package skylib

import (
	"encoding/json"
	"github.com/bketelsen/skynet/rpc/bsonrpc"
	"net/rpc"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"syscall"
)

// TODO: Better error handling, should gracefully fail to startup if it can't connect to doozer

// A Generic struct to represent any service in the SkyNet system.
type ServiceDelegate interface {
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

	RPCServ *rpc.Server   `json:"-"`
	Admin   *ServiceAdmin `json:"-"`

	Delegate ServiceDelegate `json:"-"`

	methods map[string]reflect.Value `json:"-"`
}

func CreateService(s ServiceDelegate, c *ServiceConfig) *Service {
	// This will set defaults
	initializeConfig(c)

	service := &Service{
		Config:   c,
		Delegate: s,
		Log:      c.Log,
		methods:  make(map[string]reflect.Value),
	}

	c.Log.Item(ServiceCreated{
		ServiceConfig: service.Config,
	})

	return service
}

func (s *Service) Listen(addr *BindAddr) {
	listener, err := addr.Listen()
	if err != nil {
		panic(err)
	}

	s.Log.Println(ServiceListening{
		Addr:          addr,
		ServiceConfig: s.Config,
	})

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			panic(err)
		}
		go s.RPCServ.ServeCodec(bsonrpc.NewServerCodec(conn))
	}
}

func (s *Service) Start(register bool) {

	// the main rpc server
	s.RPCServ = rpc.NewServer()
	s.RPCServ.RegisterName(s.Config.Name, s.Delegate)
	go s.Listen(s.Config.ServiceAddr)

	// the admin server
	if s.Config.AdminAddr != nil {
		s.Admin = NewServiceAdmin(s)
		go s.Admin.Listen(s.Config.AdminAddr)
	}

	// Watch signals for shutdown
	c := make(chan os.Signal, 1)
	go watchSignals(c, s)

	s.doneChan = make(chan bool, 1)
	s.Log.Item("Starting server")

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
