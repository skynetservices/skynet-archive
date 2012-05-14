package skylib

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

// A Generic struct to represent any service in the SkyNet system.
type BindAddr struct {
	IPAddress string
	Port      int
}

type ServiceInterface interface {
	Started()
	Stopped()
	Registered()
	UnRegistered()
}

type Service struct {
	ServiceAddr BindAddr
	AdminAddr   BindAddr

	Name                  string
	Region                string
	Version               string
	ConfigServers         []string
	ConfigServerDiscovery bool
	DoozerConn            *DoozerConnection
	Registered            bool
	doneChan              chan bool

	Log *log.Logger

	Delegate ServiceInterface
}

func (s *Service) Start(register bool) {
	rpc.Register(s.Delegate)
	rpc.HandleHTTP()

	portString := fmt.Sprintf("%s:%d", s.ServiceAddr.IPAddress, s.ServiceAddr.Port)

	l, e := net.Listen("tcp", portString)
	if e != nil {
		s.Log.Fatal("listen error:", e)
	}

	// Watch signals for shutdown
	c := make(chan os.Signal, 1)
	go watchSignals(c, s)

	s.doneChan = make(chan bool, 1)
	s.Log.Println("Starting server")
	go http.Serve(l, nil)

	go s.Delegate.Started() // Call user defined callback

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

	s.Delegate.Registered() // Call user defined callback
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

	s.Delegate.UnRegistered() // Call user defined callback
}

func (s *Service) Shutdown() {
	// TODO: make this wait for requests to finish
	s.UnRegister()
	s.doneChan <- true
	syscall.Exit(0)

	s.Delegate.Stopped() // Call user defined callback
}

func CreateService(s ServiceInterface, c *Config) *Service {

	// This will set defaults
	initializeConfig(c)

	service := &Service{
		Name:    c.Name,
		Version: c.Version,
		Region:  c.Region,
		ServiceAddr: BindAddr{
			IPAddress: c.IPAddress,
			Port:      c.Port,
		},
		Delegate:              s,
		Log:                   c.Log,
		ConfigServers:         c.ConfigServers,
		ConfigServerDiscovery: c.ConfigServerDiscovery,
	}

	return service
}

func initializeConfig(c *Config) {
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

	if c.IPAddress == "" {
		c.IPAddress = "127.0.0.1"
	}

	if c.Port == 0 {
		c.Port = 9000
	}

	if c.ConfigServers == nil || len(c.ConfigServers) == 0 {
		dzServers := make([]string, 0)
		dzServers = append(dzServers, "127.0.0.1:8046")
		c.ConfigServers = dzServers
	}
}

func (s *Service) GetConfigPath() string {
	return "/services/" + s.Name + "/" + s.Version + "/" + s.Region + "/" + s.ServiceAddr.IPAddress + "/" + strconv.Itoa(s.ServiceAddr.Port)
}

func (r *Service) Equal(that *Service) bool {
	var b bool
	b = false
	if r.Name != that.Name {
		return b
	}
	if r.ServiceAddr.IPAddress != that.ServiceAddr.IPAddress {
		return b
	}
	if r.ServiceAddr.Port != that.ServiceAddr.Port {
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
		s.DoozerConn = &DoozerConnection{
			Servers:  s.ConfigServers,
			Discover: s.ConfigServerDiscovery,
			Log:      s.Log,
		}

		s.DoozerConn.Connect()
	}

	return s.DoozerConn
}
