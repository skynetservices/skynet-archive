package client

import (
	"errors"
	"fmt"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/pools"
	"github.com/bketelsen/skynet/rpc/bsonrpc"
	"net"
	"net/rpc"
	"os"
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
	DoozerConn *skynet.DoozerConnection

	Config *skynet.ClientConfig
	Log    skynet.SemanticLogger `json:"-"`

	servicePools    map[string]*servicePool
	instanceMonitor *InstanceMonitor
}

func NewClient(config *skynet.ClientConfig) *Client {
	// Sanity checks (nil pointers are baaad)
	if config.Log == nil {
		config.Log = skynet.NewConsoleSemanticLogger("skynet", os.Stderr)
	}
	if config.DoozerConfig == nil {
		config.DoozerConfig = &skynet.DoozerConfig{Uri: "localhost:8046"}
	}

	if config.MaxConnectionsToInstance == 0 {
		config.Log.Fatal("Must allow at least one instance connection")
	}

	doozerConn := skynet.NewDoozerConnectionFromConfig(*config.DoozerConfig,
		config.Log)
	client := &Client{
		Config:       config,
		DoozerConn:   doozerConn,
		Log:          config.Log,
		servicePools: map[string]*servicePool{},
	}

	client.Log.Trace(fmt.Sprintf("Created client '%+v'", client))

	client.DoozerConn.Connect()

	client.instanceMonitor = NewInstanceMonitor(client.DoozerConn, false)

	return client
}

func (c *Client) doozer() *skynet.DoozerConnection {
	if c.DoozerConn == nil {
		c.DoozerConn = skynet.NewDoozerConnectionFromConfig(
			*c.Config.DoozerConfig, c.Config.Log)

		c.DoozerConn.Connect()
	}

	return c.DoozerConn
}

var servicePoolMutex sync.Mutex

func (c *Client) getServicePool(instance *skynet.ServiceInfo) (sp *servicePool) {
	servicePoolMutex.Lock()
	defer servicePoolMutex.Unlock()

	key := getInstanceKey(instance)

	var ok bool
	if sp, ok = c.servicePools[key]; ok {
		return
	}

	dbgf("making service pool, size = %d, %d\n",
		c.Config.IdleConnectionsToInstance, c.Config.MaxConnectionsToInstance)

	sp = &servicePool{
		service: instance,
		pool:    pools.NewResourcePool(getConnectionFactory(instance),
			c.Config.IdleConnectionsToInstance,
			c.Config.MaxConnectionsToInstance),
	}
	return
}

func (c *Client) GetServiceFromQuery(q *skynet.Query) (s *ServiceClient) {

	s = newServiceClient(q, c)

	return s
}

// This will not fail if no services currently exist, this saves from chicken and egg issues with dependencies between services
// TODO: We should probably determine a way of supplying secondary conditions, for example it's ok to go to a different data center only if there are no instances in our current datacenter
func (c *Client) GetService(name string, version string, region string, host string) *ServiceClient {
	registered := true
	query := &skynet.Query{
		DoozerConn: c.DoozerConn,
		Service:    name,
		Version:    version,
		Host:       host,
		Region:     region,
		Registered: &registered,
	}

	return c.GetServiceFromQuery(query)
}

func getInstanceKey(service *skynet.ServiceInfo) string {
	return service.Config.ServiceAddr.String()
}

func getConnectionFactory(s *skynet.ServiceInfo) (factory pools.Factory) {
	factory = func() (pools.Resource, error) {
		conn, err := net.Dial("tcp", s.Config.ServiceAddr.String())

		if err != nil {
			// TODO: handle failure here and attempt to connect to a different instance
			return nil, errors.New("Failed to connect to service: " + s.Config.ServiceAddr.String())
		}

		// get the service handshake
		var sh skynet.ServiceHandshake
		decoder := bsonrpc.NewDecoder(conn)

		err = decoder.Decode(&sh)
		if err != nil {
			conn.Close()
			return nil, err
		}

		ch := skynet.ClientHandshake{}
		encoder := bsonrpc.NewEncoder(conn)

		err = encoder.Encode(ch)
		if err != nil {
			conn.Close()
			return nil, err
		}

		if !sh.Registered {
			// this service has unregistered itself, look elsewhere
			conn.Close()
			return factory()
		}

		resource := ServiceResource{
			rpcClient: bsonrpc.NewClient(conn),
			service:   s,
			clientID:  sh.ClientID,
		}

		return resource, nil
	}
	return
}
