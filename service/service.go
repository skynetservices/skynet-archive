package service

import (
	"encoding/json"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/rpc/bsonrpc"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"sync"
	"syscall"
	"time"
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
	DoozerConn *skynet.DoozerConnection `json:"-"`
	skynet.ServiceInfo

	doneChan chan bool `json:"-"`

	Log skynet.Logger `json:"-"`

	RPCServ *rpc.Server   `json:"-"`
	Admin   *ServiceAdmin `json:"-"`

	Delegate ServiceDelegate `json:"-"`

	methods map[string]reflect.Value `json:"-"`

	activeRequests sync.WaitGroup `json:"-"`

	connectionChan chan *net.TCPConn `json:"-"`
	registeredChan chan bool         `json:"-"`

	doozerChan   chan interface{} `json:"-"`
	doozerWaiter sync.WaitGroup   `json:"-"`

	rpcListener  *net.TCPListener `json:"-"`
	updateTicker *time.Ticker     `json:"-"`
}

func CreateService(sd ServiceDelegate, c *skynet.ServiceConfig) (s *Service) {
	// This will set defaults
	initializeConfig(c)

	s = &Service{
		Delegate:       sd,
		Log:            c.Log,
		methods:        make(map[string]reflect.Value),
		connectionChan: make(chan *net.TCPConn),
		registeredChan: make(chan bool),
		doozerChan:     make(chan interface{}),
		updateTicker:   time.NewTicker(c.DoozerUpdateInterval),
	}

	s.Config = c
	s.Stats = skynet.ServiceStatistics{
		StartTime: time.Now().Format("2006-01-02T15:04:05Z-0700"),
	}

	c.Log.Item(skynet.ServiceCreated{
		ServiceConfig: s.Config,
	})

	// the main rpc server
	s.RPCServ = rpc.NewServer()
	rpcForwarder := NewServiceRPC(s)

	c.Log.Item(RegisteredMethods{rpcForwarder.MethodNames})

	s.RPCServ.RegisterName(s.Config.Name, rpcForwarder)

	return
}

func (s *Service) listen(addr *skynet.BindAddr) {
	var err error
	s.rpcListener, err = addr.Listen()
	if err != nil {
		panic(err)
	}

	s.Log.Item(skynet.ServiceListening{
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
			s.Stats.Clients += 1

			// send the server handshake
			sh := skynet.ServiceHandshake{
				Registered: s.Registered,
			}
			encoder := bsonrpc.NewEncoder(conn)
			err := encoder.Encode(sh)
			if err != nil {
				s.Log.Item(err)

				s.Stats.Clients -= 1
				break
			}
			if !s.Registered {
				conn.Close()
				s.Stats.Clients -= 1
				break
			}

			// read the client handshake
			var ch skynet.ClientHandshake
			decoder := bsonrpc.NewDecoder(conn)
			err = decoder.Decode(&ch)
			if err != nil {
				s.Log.Item(err)
				s.Stats.Clients -= 1
				break
			}

			// here do stuff with the client handshake

			go func() {
				s.RPCServ.ServeCodec(bsonrpc.NewServerCodec(conn))

				s.Stats.Clients -= 1
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

		case _ = <-s.updateTicker.C:
			s.UpdateCluster()
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
func (s *Service) doozer() *skynet.DoozerConnection {
	if s.DoozerConn == nil {
		s.DoozerConn = skynet.NewDoozerConnectionFromConfig(*s.Config.DoozerConfig, s.Log)

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

	// If doozer contains instances with the same ip:port we just bound to then they are no longer alive and need to be cleaned up
	s.cleanupDoozerEntriesForAddr(s.Config.ServiceAddr.IPAddress, s.Config.ServiceAddr.Port)
	s.cleanupDoozerEntriesForAddr(s.Config.AdminAddr.IPAddress, s.Config.AdminAddr.Port)

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

func (s *Service) cleanupDoozerEntriesForAddr(ip string, port int) {
	q := skynet.Query{
		Host:       ip,
		Port:       strconv.Itoa(port),
		DoozerConn: s.doozer(),
	}

	instances := q.FindInstances()

	for _, i := range instances {
		s.Log.Item("Cleaning up old doozer entry with conflicting addr " + ip + ":" + strconv.Itoa(port) + "(" + i.GetConfigPath() + ")")
		s.doozer().Del(i.GetConfigPath(), s.doozer().GetCurrentRevision())
	}
}

func (s *Service) UpdateCluster() {
	b, err := json.Marshal(s.ServiceInfo)
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
	s.doneChan <- true

	s.activeRequests.Wait()
	s.doozerWaiter.Wait()

	s.Delegate.Stopped(s) // Call user defined callback

}

func initializeConfig(c *skynet.ServiceConfig) {
	if c.Log == nil {
		c.Log = skynet.NewConsoleLogger(os.Stderr)
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
		c.DoozerConfig = &skynet.DoozerConfig{
			Uri:          "127.0.0.1:8046",
			AutoDiscover: true,
		}
	}

	if c.DoozerUpdateInterval == 0 {
		c.DoozerUpdateInterval = 5 * time.Second
	}
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
