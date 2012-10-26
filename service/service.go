package service

import (
	"encoding/json"
	"fmt"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/rpc/bsonrpc"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
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

type ClientInfo struct {
	Address net.Addr
}

type Service struct {
	DoozerConn *skynet.DoozerConnection
	skynet.ServiceInfo

	// for sending the signal into mux()
	doneChan chan bool
	// for waiting for all shutdown operations
	doneGroup *sync.WaitGroup

	Log skynet.SemanticLogger

	RPCServ *rpc.Server
	Admin   *ServiceAdmin

	Delegate ServiceDelegate

	methods map[string]reflect.Value

	activeRequests sync.WaitGroup

	connectionChan chan *net.TCPConn
	registeredChan chan bool

	doozerChan   chan interface{}
	doozerWaiter sync.WaitGroup

	rpcListener  *net.TCPListener
	updateTicker *time.Ticker

	clientMutex sync.Mutex
	clientInfo  map[string]ClientInfo
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
		clientInfo:     make(map[string]ClientInfo),
	}

	s.Config = c
	s.Stats = &skynet.ServiceStatistics{
		StartTime: time.Now().Format("2006-01-02T15:04:05Z-0700"),
	}

	c.Log.Trace(fmt.Sprintf("%+v", skynet.ServiceCreated{
		ServiceConfig: s.Config,
	}))

	// the main rpc server
	s.RPCServ = rpc.NewServer()
	rpcForwarder := NewServiceRPC(s)

	c.Log.Trace(fmt.Sprintf("%+v", RegisteredMethods{rpcForwarder.MethodNames}))

	s.RPCServ.RegisterName(s.Config.Name, rpcForwarder)

	return
}

func (s *Service) listen(addr *skynet.BindAddr, bindWait *sync.WaitGroup) {
	var err error
	s.rpcListener, err = addr.Listen()
	if err != nil {
		panic(err)
	}

	s.Log.Trace(fmt.Sprintf("%+v", ServiceListening{
		Addr:          addr,
		ServiceConfig: s.Config,
	}))

	bindWait.Done()

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
			atomic.AddInt32(&s.Stats.Clients, 1)

			clientID := skynet.UUID()

			s.clientMutex.Lock()
			s.clientInfo[clientID] = ClientInfo{
				Address: conn.RemoteAddr(),
			}
			s.clientMutex.Unlock()

			// send the server handshake
			sh := skynet.ServiceHandshake{
				Registered: s.Registered,
				ClientID:   clientID,
			}
			encoder := bsonrpc.NewEncoder(conn)
			err := encoder.Encode(sh)
			if err != nil {
				s.Log.Error(err.Error())

				atomic.AddInt32(&s.Stats.Clients, -1)
				break
			}
			if !s.Registered {
				conn.Close()
				atomic.AddInt32(&s.Stats.Clients, -1)
				break
			}

			// read the client handshake
			var ch skynet.ClientHandshake
			decoder := bsonrpc.NewDecoder(conn)
			err = decoder.Decode(&ch)
			if err != nil {
				s.Log.Error("Error calling bsonrpc.NewDecoder: " + err.Error())
				atomic.AddInt32(&s.Stats.Clients, -1)
				break
			}

			// here do stuff with the client handshake

			go func() {
				s.RPCServ.ServeCodec(bsonrpc.NewServerCodec(conn))

				atomic.AddInt32(&s.Stats.Clients, -1)
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
			s.UpdateDoozerStats()
		}
	}
}

type doozerSetConfig struct {
	ConfigPath string
	ConfigData []byte
}

type doozerRemoveFromCluster struct {
	ConfigPath string
	StatsPath  string
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
				s.Log.Fatal(err.Error())
			}
		case doozerRemoveFromCluster:
			rev := s.doozer().GetCurrentRevision()
			err := s.doozer().Del(i.ConfigPath, rev)
			if err != nil {
				s.Log.Fatal(err.Error())
			}

			err = s.doozer().Del(i.StatsPath, rev)
			if err != nil {
				s.Log.Fatal(err.Error())
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

func (s *Service) getClientInfo(clientID string) (ci ClientInfo, ok bool) {
	s.clientMutex.Lock()
	defer s.clientMutex.Unlock()

	ci, ok = s.clientInfo[clientID]
	return
}

func (s *Service) IsTrusted(addr net.Addr) bool {
	// TODO: something else
	return false
}

func (s *Service) Start(register bool) (done *sync.WaitGroup) {
	bindWait := &sync.WaitGroup{}

	bindWait.Add(1)
	go s.listen(s.Config.ServiceAddr, bindWait)

	// the admin server
	if s.Config.AdminAddr != nil {
		s.Admin = NewServiceAdmin(s)
		bindWait.Add(1)
		go s.Admin.Listen(s.Config.AdminAddr, bindWait)
	} else {
		s.Log.Trace(fmt.Sprintf("%+v", AdminNotListening{s.Config}))
	}

	// Watch signals for shutdown
	c := make(chan os.Signal, 1)
	go watchSignals(c, s)

	s.doneChan = make(chan bool, 1)

	// We must block here, we don't want to register with doozer, until we've actually bound to an ip:port
	bindWait.Wait()

	// If doozer contains instances with the same ip:port we just bound to then they are no longer alive and need to be cleaned up
	s.cleanupDoozerEntriesForAddr(s.Config.ServiceAddr)
	s.cleanupDoozerEntriesForAddr(s.Config.AdminAddr)

	go s.Delegate.Started(s) // Call user defined callback

	s.doozerWaiter.Add(1)
	go s.doozerMux()

	s.UpdateDoozerServiceInfo()

	if register == true {
		s.register()
	}

	s.doneGroup = &sync.WaitGroup{}
	s.doneGroup.Add(1)
	go func() {
		s.mux()
		s.doneGroup.Done()
	}()
	done = s.doneGroup
	return
}

func (s *Service) cleanupDoozerEntriesForAddr(addr *skynet.BindAddr) {
	if addr == nil {
		return
	}
	q := skynet.Query{
		Host:       addr.IPAddress,
		Port:       strconv.Itoa(addr.Port),
		DoozerConn: s.doozer(),
	}

	instances := q.FindInstances()

	for _, i := range instances {
		s.Log.Trace("Cleaning up old doozer entry with conflicting addr " +
			addr.String() + "(" + i.GetConfigPath() + ")")
		s.doozer().Del(i.GetConfigPath(), s.doozer().GetCurrentRevision())
		s.doozer().Del(i.GetStatsPath(), s.doozer().GetCurrentRevision())
	}
}

func (s *Service) UpdateDoozerServiceInfo() {

	// We're going to create a copy of our ServiceInfo so that we can nil out the Stats, which will match the omitempty and won't marshal 
	// this is cheap as it's a single bool, and 2 pointers.
	si := s.ServiceInfo
	si.Stats = nil

	b, err := json.Marshal(si)
	if err != nil {
		s.Log.Fatal(err.Error())
	}
	cfgpath := s.GetConfigPath()

	s.doozerChan <- doozerSetConfig{
		ConfigPath: cfgpath,
		ConfigData: b,
	}
}

func (s *Service) UpdateDoozerStats() {
	b, err := json.Marshal(s.ServiceInfo.Stats)
	if err != nil {
		s.Log.Fatal(err.Error())
	}
	s.doozerChan <- doozerSetConfig{
		ConfigPath: s.GetStatsPath(),
		ConfigData: b,
	}
}

func (s *Service) RemoveFromCluster() {
	s.doozerChan <- doozerRemoveFromCluster{
		ConfigPath: s.GetConfigPath(),
		StatsPath:  s.GetStatsPath(),
	}
}

func (s *Service) register() {
	// this version must be run from the runService() goroutine
	if s.Registered {
		return
	}
	s.Registered = true
	s.Log.Trace(fmt.Sprintf("%+v", ServiceRegistered{s.Config}))
	s.UpdateDoozerServiceInfo()
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
	s.Log.Trace(fmt.Sprintf("%+v", ServiceUnregistered{s.Config}))
	s.UpdateDoozerServiceInfo()
	s.Delegate.Unregistered(s) // Call user defined callback
}

func (s *Service) Unregister() {
	s.registeredChan <- false
}

func (s *Service) Shutdown() {
	s.doneGroup.Add(1)

	s.doneChan <- true

	s.activeRequests.Wait()
	s.doozerWaiter.Wait()

        if s.Delegate != nil {
	    s.Delegate.Stopped(s) // Call user defined callback
            s.Delegate = nil
        }

	s.doneGroup.Done()
}

func initializeConfig(c *skynet.ServiceConfig) {
	if c.Log == nil {
		c.Log = skynet.NewConsoleSemanticLogger("skynet", os.Stderr)
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

	if c.ServiceAddr == nil {
		c.ServiceAddr = &skynet.BindAddr{}
	}

	if c.ServiceAddr.IPAddress == "" {
		c.ServiceAddr.IPAddress = "127.0.0.1"
	}

	if c.ServiceAddr.Port == 0 {
		c.ServiceAddr.Port = 9000
	}
	if c.ServiceAddr.MaxPort == 0 {
		c.ServiceAddr.MaxPort = 9999
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
			case syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT,
				syscall.SIGSEGV, syscall.SIGSTOP, syscall.SIGTERM:
				s.Log.Trace(fmt.Sprintf("%+v", KillSignal{sig.(syscall.Signal)}))
				s.Shutdown()
			}
		}
	}
}
