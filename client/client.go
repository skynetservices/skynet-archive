package client

import (
	"errors"
	"github.com/skynetservices/skynet2"
	"github.com/skynetservices/skynet2/log"
	"github.com/skynetservices/skynet2/pools"
	"github.com/skynetservices/skynet2/rpc/bsonrpc"
	"net"
	"net/rpc"
	"sync"
)

var (
	ErrServiceUnregistered = errors.New("Service is unregistered")
	ErrRequestTimeout      = errors.New("Service request timed out")
)

type ServiceResource struct {
	rpcClient *rpc.Client
	service   *skynet.ServiceInfo
	clientID  string
	closed    bool
}

func (s ServiceResource) Close() {
	s.closed = true
	s.rpcClient.Close()
}

func (s ServiceResource) IsClosed() bool {
	return s.closed
}

type Client struct {
	Config       *skynet.ClientConfig
	servicePools map[string]*servicePool
}

func NewClient(config *skynet.ClientConfig) *Client {
	if config.MaxConnectionsToInstance == 0 {
		log.Fatal("Must allow at least one instance connection")
	}

	client := &Client{
		Config:       config,
		servicePools: map[string]*servicePool{},
	}

	return client
}

var servicePoolMutex sync.Mutex

func (c *Client) getServicePool(instance *skynet.ServiceInfo) (sp *servicePool) {
	log.Println(log.TRACE, "in getServicePool")

	servicePoolMutex.Lock()
	log.Println(log.TRACE, "have servicePoolMutex lock")
	defer servicePoolMutex.Unlock()

	key := getInstanceKey(instance)

	var ok bool
	if sp, ok = c.servicePools[key]; ok {
		log.Println(log.TRACE, "returning existing connection from servicePool")
		return
	}

	log.Println(log.TRACE, "creating new connection in servicePool")
	sp = &servicePool{
		service: instance,
		pool: pools.NewResourcePool(getConnectionFactory(instance),
			c.Config.IdleConnectionsToInstance,
			c.Config.MaxConnectionsToInstance),
	}
	return
}

func (c *Client) removeServicePool(instance skynet.ServiceInfo) {
	servicePoolMutex.Lock()
	defer servicePoolMutex.Unlock()

	delete(c.servicePools, getInstanceKey(&instance))
}

func (c *Client) GetService(criteria *skynet.Criteria) *ServiceClient {
	return newServiceClient(criteria, c)
}

func getInstanceKey(service *skynet.ServiceInfo) string {
	return service.ServiceAddr.String()
}

func getConnectionFactory(s *skynet.ServiceInfo) (factory pools.Factory) {
	factory = func() (pools.Resource, error) {
		log.Println(log.DEBUG, "Attempting connection to: "+s.ServiceAddr.String())
		conn, err := net.Dial("tcp", s.ServiceAddr.String())
		log.Println(log.DEBUG, "dial returned")

		if err != nil {
			// TODO: handle failure here and attempt to connect to a different instance
			log.Println(log.ERROR, "Failed connection attempt to instance", err)
			return nil, errors.New("Failed to connect to service: " + s.ServiceAddr.String())
		}

		// get the service handshake
		var sh skynet.ServiceHandshake
		decoder := bsonrpc.NewDecoder(conn)

		log.Println(log.DEBUG, "Decoding ServiceHandshake")
		err = decoder.Decode(&sh)
		if err != nil {
			log.Println(log.ERROR, "Failed to decode ServiceHandshake", err)
			conn.Close()
			return nil, err
		}

		log.Println(log.DEBUG, sh)

		ch := skynet.ClientHandshake{}
		encoder := bsonrpc.NewEncoder(conn)

		log.Println(log.DEBUG, "Encoding ClientHandshake")
		err = encoder.Encode(ch)
		if err != nil {
			log.Println(log.ERROR, "Failed to encode ClientHandshake", err)
			conn.Close()
			return nil, err
		}

		log.Println(log.DEBUG, "Checking if service is registered")
		if !sh.Registered {
			log.Println(log.DEBUG, "Instance not registered")
			// this service has unregistered itself, look elsewhere
			conn.Close()
			return factory()
		}

		log.Println(log.DEBUG, "returning resource")
		resource := ServiceResource{
			rpcClient: bsonrpc.NewClient(conn),
			service:   s,
			clientID:  sh.ClientID,
		}

		return resource, nil
	}
	return
}
