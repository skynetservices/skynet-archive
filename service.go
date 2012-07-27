package skynet

import (
	"encoding/json"
	"github.com/bketelsen/skynet/rpc/bsonrpc"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"sync"
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

	activeClients sync.WaitGroup `json:"-"`

	connectionChan chan *net.TCPConn `json:"-"`
	registeredChan chan bool         `json:"-"`

	doozerChan   chan interface{} `json:"-"`
	doozerWaiter sync.WaitGroup   `json:"-"`

	rpcListener *net.TCPListener `json:"-"`
}

func CreateService(sd ServiceDelegate, c *ServiceConfig) (s *Service) {
	// This will set defaults
	initializeConfig(c)

	s = &Service{
		Config:         c,
		Delegate:       sd,
		Log:            c.Log,
		methods:        make(map[string]reflect.Value),
		connectionChan: make(chan *net.TCPConn),
		registeredChan: make(chan bool),
		doozerChan:     make(chan interface{}),
	}

	c.Log.Item(ServiceCreated{
		ServiceConfig: s.Config,
	})
	// the main rpc server
	s.RPCServ = rpc.NewServer()
	rpcForwarder := NewServiceRPC(s.Delegate, c.Log)

	c.Log.Item(RegisteredMethods{rpcForwarder.MethodNames})

	s.RPCServ.RegisterName(s.Config.Name, rpcForwarder)

	return
}

func (s *Service) listen(addr *BindAddr) {
	var err error
	s.rpcListener, err = addr.Listen()
	if err != nil {
		panic(err)
	}

	s.Log.Item(ServiceListening{
		Addr:          addr,
		ServiceConfig: s.Config,
	})

	for {
		conn, err := s.rpcListener.AcceptTCP()
		if err != nil {
			panic(err)
		}
		s.connectionChan <- conn
	}
}

// this function is the goroutine that owns this service - all thread-sensitive data needs to
// be manipulated only through here.
func (s *Service) mux() {
loop:
	for {
		select {
		case conn := <-s.connectionChan:
			s.activeClients.Add(1)

			// send the server handshake
			sh := ServiceHandshake{
				Registered: s.Registered,
			}
			encoder := bsonrpc.NewEncoder(conn)
			err := encoder.Encode(sh)
			if err != nil {
				s.Log.Item(err)
				s.activeClients.Done()
				break
			}
			if !s.Registered {
				conn.Close()
				s.activeClients.Done()
				break
			}

			// read the client handshake
			var ch ClientHandshake
			decoder := bsonrpc.NewDecoder(conn)
			err = decoder.Decode(&ch)
			if err != nil {
				s.Log.Item(err)
				s.activeClients.Done()
				break
			}

			// here do stuff with the client handshake

			go func() {
				s.RPCServ.ServeCodec(bsonrpc.NewServerCodec(conn))
				s.activeClients.Done()
			}()
		case register := <-s.registeredChan:
			if register {
				s.register()
			} else {
				s.unregister()
			}
		case _ = <-s.doneChan:
			go func() {
				for _ = range s.doneChan {
				}
			}()
			s.RemoveFromCluster()
			s.doozerChan <- doozerFinish{}
			break loop
		}
	}
}

type doozerSetConfig struct {
	ConfigPath string
	ConfigData []byte
}

type doozerRemoveFromCluster struct {
	ConfigPath string
}

type doozerFinish struct{}

func (s *Service) doozerMux() {
loop:
	for i := range s.doozerChan {
		switch i := i.(type) {
		case doozerSetConfig:
			rev := s.doozer().GetCurrentRevision()
			_, err := s.DoozerConn.Set(i.ConfigPath, rev, i.ConfigData)
			if err != nil {
				s.Log.Panic(err.Error())
			}
		case doozerRemoveFromCluster:
			rev := s.doozer().GetCurrentRevision()
			err := s.doozer().Del(i.ConfigPath, rev)
			if err != nil {
				s.Log.Panic(err.Error())
			}
		case doozerFinish:
			break loop
		}
	}
	s.doozerWaiter.Done()
}

// only call this from doozerMux
func (s *Service) doozer() DoozerConnection {
	if s.DoozerConn == nil {
		s.DoozerConn = NewDoozerConnectionFromConfig(*s.Config.DoozerConfig, s.Log)

		s.DoozerConn.Connect()
	}

	return s.DoozerConn
}

func (s *Service) Start(register bool) (done *sync.WaitGroup) {

	go s.listen(s.Config.ServiceAddr)

	// the admin server
	if s.Config.AdminAddr != nil {
		s.Admin = NewServiceAdmin(s)
		go s.Admin.Listen(s.Config.AdminAddr)
	}

	// Watch signals for shutdown
	c := make(chan os.Signal, 1)
	go watchSignals(c, s)

	s.doneChan = make(chan bool, 1)

	go s.Delegate.Started(s) // Call user defined callback

	s.doozerWaiter.Add(1)
	go s.doozerMux()

	s.UpdateCluster()

	if register == true {
		s.register()
	}

	done = &sync.WaitGroup{}
	done.Add(1)
	go func() {
		s.mux()
		done.Done()
	}()
	return
}

func (s *Service) UpdateCluster() {
	b, err := json.Marshal(s)
	if err != nil {
		s.Log.Panic(err.Error())
	}
	cfgpath := s.GetConfigPath()

	s.doozerChan <- doozerSetConfig{
		ConfigPath: cfgpath,
		ConfigData: b,
	}
}

func (s *Service) RemoveFromCluster() {
	s.doozerChan <- doozerRemoveFromCluster{
		ConfigPath: s.GetConfigPath(),
	}
}

func (s *Service) register() {
	// this version must be run from the runService() goroutine
	if s.Registered {
		return
	}
	s.Registered = true
	s.UpdateCluster()
	s.Delegate.Registered(s) // Call user defined callback
}

func (s *Service) Register() {
	s.registeredChan <- true
}

func (s *Service) unregister() {
	// this version must be run from the runService() goroutine
	if !s.Registered {
		return
	}
	s.Registered = false
	s.UpdateCluster()
	s.Delegate.Unregistered(s) // Call user defined callback
}

func (s *Service) Unregister() {
	s.registeredChan <- false
}

func (s *Service) Shutdown() {

	// TODO: make this wait for requests to finish
	s.doneChan <- true

	s.activeClients.Wait()
	s.doozerWaiter.Wait()

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
	signal.Notify(c, syscall.SIGINT, syscall.SIGKILL, syscall.SIGSEGV, syscall.SIGSTOP, syscall.SIGTERM)

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
