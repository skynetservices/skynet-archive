package service

import (
	"github.com/skynetservices/skynet2"
	"github.com/skynetservices/skynet2/daemon"
	"github.com/skynetservices/skynet2/log"
	"github.com/skynetservices/skynet2/rpc/bsonrpc"
	"io"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
)

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
	skynet.ServiceInfo
	Delegate       ServiceDelegate
	methods        map[string]reflect.Value
	RPCServ        *rpc.Server
	rpcListener    *net.TCPListener
	activeRequests sync.WaitGroup
	connectionChan chan *net.TCPConn
	registeredChan chan bool
	shutdownChan   chan bool
	listenChan     chan bool

	clientMutex sync.Mutex
	ClientInfo  map[string]ClientInfo

	// for sending the signal into mux()
	doneChan chan bool

	// for waiting for all shutdown operations
	doneGroup *sync.WaitGroup

	shuttingDown bool
	pipe         *daemon.Pipe
}

// Wraps your custom service in Skynet
func CreateService(sd ServiceDelegate, c *skynet.ServiceConfig) (s *Service) {
	s = &Service{
		Delegate:       sd,
		methods:        make(map[string]reflect.Value),
		connectionChan: make(chan *net.TCPConn),
		registeredChan: make(chan bool),
		shutdownChan:   make(chan bool),
		listenChan:     make(chan bool),
		ClientInfo:     make(map[string]ClientInfo),
		shuttingDown:   false,
	}

	s.ServiceConfig = c

	// the main rpc server
	s.RPCServ = rpc.NewServer()
	rpcForwarder := NewServiceRPC(s)
	s.RPCServ.RegisterName(s.ServiceConfig.Name, rpcForwarder)

	// Daemon doesn't accept commands over pipe
	if c.Name != "SkynetDaemon" {
		// Listen for admin requests
		go s.serveAdminRequests()
	}

	return
}

// Notifies the cluster your service is ready to handle requests
func (s *Service) Register() {
	s.registeredChan <- true
}

func (s *Service) register() {
	// this version must be run from the mux() goroutine
	if s.Registered {
		return
	}

	err := skynet.GetServiceManager().Register(s.ServiceConfig.UUID)
	if err != nil {
		log.Println(log.ERROR, "Failed to register service: "+err.Error())
	}

	s.Registered = true
	log.Printf(log.INFO, "%+v\n", ServiceRegistered{s.ServiceConfig})
	s.Delegate.Registered(s) // Call user defined callback
}

// Leave your service online, but notify the cluster it's not currently accepting new requests
func (s *Service) Unregister() {
	s.registeredChan <- false
}

func (s *Service) unregister() {
	// this version must be run from the mux() goroutine
	if !s.Registered {
		return
	}

	err := skynet.GetServiceManager().Unregister(s.ServiceConfig.UUID)
	if err != nil {
		log.Println(log.ERROR, "Failed to unregister service: "+err.Error())
	}

	s.Registered = false
	log.Printf(log.INFO, "%+v\n", ServiceUnregistered{s.ServiceConfig})
	s.Delegate.Unregistered(s) // Call user defined callback
}

func (s *Service) Shutdown() {
	if s.shuttingDown {
		return
	}

	s.registeredChan <- false
	s.shutdownChan <- true
}

// Wait for existing requests to complete and shutdown service
func (s *Service) shutdown() {
	if s.shuttingDown {
		return
	}

	s.shuttingDown = true

	s.doneGroup.Add(1)
	s.rpcListener.Close()

	s.doneChan <- true

	s.activeRequests.Wait()

	err := skynet.GetServiceManager().Remove(s.ServiceInfo)
	if err != nil {
		log.Println(log.ERROR, "Failed to remove service: "+err.Error())
	}

	s.Delegate.Stopped(s) // Call user defined callback

	s.doneGroup.Done()
}

// TODO: Currently unimplemented
func (s *Service) IsTrusted(addr net.Addr) bool {
	return false
}

// Starts your skynet service, including binding to ports. Optionally register for requests at the same time. Returns a sync.WaitGroup that will block until all requests have finished
func (s *Service) Start(register bool) (done *sync.WaitGroup) {
	bindWait := &sync.WaitGroup{}

	bindWait.Add(1)
	go s.listen(s.ServiceConfig.ServiceAddr, bindWait)

	// Watch signals for shutdown
	c := make(chan os.Signal, 1)
	go watchSignals(c, s)

	s.doneChan = make(chan bool, 1)

	// We must block here, we don't want to register, until we've actually bound to an ip:port
	bindWait.Wait()

	s.doneGroup = &sync.WaitGroup{}
	s.doneGroup.Add(1)

	go func() {
		s.mux()
		s.doneGroup.Done()
	}()
	done = s.doneGroup

	err := skynet.GetServiceManager().Add(s.ServiceInfo)
	if err != nil {
		log.Println(log.ERROR, "Failed to add service: "+err.Error())
	}

	if register {
		s.Register()
	}

	go s.Delegate.Started(s) // Call user defined callback

	return
}

func (s *Service) getClientInfo(clientID string) (ci ClientInfo, ok bool) {
	s.clientMutex.Lock()
	defer s.clientMutex.Unlock()

	ci, ok = s.ClientInfo[clientID]
	return
}

func (s *Service) listen(addr skynet.BindAddr, bindWait *sync.WaitGroup) {
	var err error
	s.rpcListener, err = addr.Listen()
	if err != nil {
		panic(err)
	}

	log.Printf(log.INFO, "%+v\n", ServiceListening{
		Addr:          &addr,
		ServiceConfig: s.ServiceConfig,
	})

	// We may have changed port due to conflict, ensure config has the correct port now
	a, _ := skynet.BindAddrFromString(addr.String())
	s.ServiceConfig.ServiceAddr.IPAddress = a.IPAddress
	s.ServiceConfig.ServiceAddr.Port = a.Port

	bindWait.Done()

	for {
		conn, err := s.rpcListener.AcceptTCP()
		if err != nil && !s.shuttingDown {
			log.Println(log.ERROR, "AcceptTCP failed", err)
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
			clientID := skynet.UUID()

			s.clientMutex.Lock()
			s.ClientInfo[clientID] = ClientInfo{
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
				log.Println(log.ERROR, err.Error())
				break
			}
			if !s.Registered {
				conn.Close()
				break
			}

			// read the client handshake
			var ch skynet.ClientHandshake
			decoder := bsonrpc.NewDecoder(conn)
			err = decoder.Decode(&ch)
			if err != nil {
				log.Println(log.ERROR, "Error calling bsonrpc.NewDecoder: "+err.Error())
				break
			}

			// here do stuff with the client handshake
			go func() {
				s.RPCServ.ServeCodec(bsonrpc.NewServerCodec(conn))
			}()
		case register := <-s.registeredChan:
			if register {
				s.register()
			} else {
				s.unregister()
			}
		case <-s.shutdownChan:
			s.shutdown()
		case _ = <-s.doneChan:
			break loop
		}
	}
}

func (s *Service) serveAdminRequests() {
	pipeReader := os.NewFile(uintptr(os.Stderr.Fd()+1), "")
	pipeWriter := os.NewFile(uintptr(os.Stderr.Fd()+2), "")
	s.pipe = daemon.NewPipe(pipeReader, pipeWriter)

	b := make([]byte, daemon.MAX_PIPE_BYTES)
	for {
		n, err := s.pipe.Read(b)

		if err != nil && err != io.EOF {
			log.Printf(log.ERROR, "Error reading from admin pipe "+err.Error())
			return
		}

		cmd := string(b[:n])
		log.Println(log.TRACE, "Received "+cmd+" from daemon")

		switch cmd {
		case "SHUTDOWN":
			s.Shutdown()
			s.pipe.Write([]byte("ACK"))
			break
		case "REGISTER":
			s.Register()
			s.pipe.Write([]byte("ACK"))
		case "UNREGISTER":
			s.Unregister()
			s.pipe.Write([]byte("ACK"))
		}
	}
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
				log.Printf(log.INFO, "%+v", KillSignal{sig.(syscall.Signal)})
				s.Shutdown()
				return
			}
		}
	}
}
