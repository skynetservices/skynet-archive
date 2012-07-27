package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/rpc/bsonrpc"
	"github.com/bketelsen/skynet/service"
	"github.com/bketelsen/skynet/util"
	"launchpad.net/mgo/v2/bson"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"strings"
)

var (
	ErrServiceUnregistered = errors.New("Service is unregistered")
)

type Client struct {
	DoozerConn skynet.DoozerConnection

	Config *skynet.ClientConfig
	Log    skynet.Logger `json:"-"`
}

type ServiceResource struct {
	rpcClient *rpc.Client
	service   service.Service
	closed    bool
}

func (s ServiceResource) Close() {
	s.closed = true
	s.rpcClient.Close()
}

func (s ServiceResource) IsClosed() bool {
	return s.closed
}

func (c *Client) doozer() skynet.DoozerConnection {
	if c.DoozerConn == nil {
		c.DoozerConn = skynet.NewDoozerConnectionFromConfig(*c.Config.DoozerConfig, c.Config.Log)

		c.DoozerConn.Connect()
	}

	return c.DoozerConn
}

func NewClient(config *skynet.ClientConfig) *Client {
	if config.Log == nil {
		config.Log = skynet.NewConsoleLogger(os.Stderr)
	}

	if config.ConnectionPoolSize == 0 {
		config.ConnectionPoolSize = 1
	}

	client := &Client{
		Config:     config,
		DoozerConn: skynet.NewDoozerConnectionFromConfig(*config.DoozerConfig, config.Log),
		Log:        config.Log,
	}

	client.Log.Item(config)

	client.DoozerConn.Connect()

	return client
}

func (c *Client) GetServiceFromQuery(q *Query) (s *ServiceClient) {
	var conn net.Conn
	var err error

	s = &ServiceClient{
		Log:            c.Config.Log,
		connectionPool: pools.NewRoundRobin(c.Config.ConnectionPoolSize, c.Config.IdleTimeout),
		query:          q,
		instances:      make(map[string]service.Service, 0),
	}

	// Load initial list of instances
	results := s.query.FindInstances()

	if results != nil {
		for _, instance := range results {
			key := instance.Config.ServiceAddr.IPAddress + ":" + strconv.Itoa(instance.Config.ServiceAddr.Port)
			s.instances[key] = *instance
		}
	}

	go s.monitorInstances()

	var factory func() (pools.Resource, error)
	factory = func() (pools.Resource, error) {
		if len(s.instances) < 1 {

			return nil, errors.New("No services available that match your criteria")
		}

		// Connect to random instance
		index := (rand.Int() % len(s.instances))

		var instance service.Service

		i := 0

		var key string
		for k, v := range s.instances {
			if i == index {
				key = k
				instance = v
				break
			}
		}

		conn, err = net.Dial("tcp", instance.Config.ServiceAddr.String())

		if err != nil {
			// TODO: handle failure here and attempt to connect to a different instance
			return nil, errors.New("Failed to connect to service: " + instance.Config.ServiceAddr.String())
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
			delete(s.instances, key)
			return factory()
		}

		resource := ServiceResource{
			rpcClient: bsonrpc.NewClient(conn),
			service:   instance,
		}

		return resource, nil
	}

	s.connectionPool.Open(factory)

	return s
}

// This will not fail if no services currently exist, this saves from chicken and egg issues with dependencies between services
// TODO: We should probably determine a way of supplying secondary conditions, for example it's ok to go to a different data center only if there are no instances in our current datacenter
func (c *Client) GetService(name string, version string, region string, host string) *ServiceClient {
	registered := true
	query := &Query{
		DoozerConn: c.DoozerConn,
		Service:    name,
		Version:    version,
		Host:       host,
		Region:     region,
		Registered: &registered,
	}

	return c.GetServiceFromQuery(query)
}

type ServiceClient struct {
	Log            skynet.Logger `json:"-"`
	connectionPool *pools.RoundRobin
	query          *Query
	instances      map[string]service.Service
}

func (c *ServiceClient) monitorInstances() {
	// TODO: Let's watch doozer and keep this list up to date so we don't need to search it every time we spawn a new connection
	doozer := c.query.DoozerConn

	rev := doozer.GetCurrentRevision()

	for {
		ev, err := doozer.Wait("/services/**", rev+1)
		rev = ev.Rev

		if err == nil {
			var s service.Service

			buf := bytes.NewBuffer(ev.Body)

			err = json.Unmarshal(buf.Bytes(), &s)

			if err == nil {
				parts := strings.Split(ev.Path, "/")

				if c.query.pathMatches(parts, ev.Path) {
					key := s.Config.ServiceAddr.String()

					if s.Registered == true {
						//c.Log.Println("New Service Instance Discovered: " + key)
						c.Log.Item(service.ServiceDiscovered{
							Service: &s,
						})
						c.instances[key] = s
					} else {
						//c.Log.Println("Service Instance Removed: " + key)
						c.Log.Item(service.ServiceRemoved{
							Service: &s,
						})
						delete(c.instances, key)
					}
				}
			}
		}
	}
}

func (c *ServiceClient) Send(requestInfo *skynet.RequestInfo, funcName string, in interface{}, outPointer interface{}) (err error) {
	// TODO: timeout logic
	s, err := c.getConnection(0)
	if err != nil {
		c.Log.Item(err)
		return
	}

	if requestInfo == nil {
		requestInfo = &skynet.RequestInfo{
			RequestID: skynet.UUID(),
		}
	}

	sin := service.ServiceRPCIn{
		RequestInfo: requestInfo,
		Method:      funcName,
	}

	sin.In, err = bson.Marshal(in)
	if err != nil {
		return
	}

	sout := service.ServiceRPCOut{}

	// TODO: Check for connectivity issue so that we can try to get another resource out of the pool
	err = s.rpcClient.Call(s.service.Config.Name+".Forward", sin, &sout)
	if err != nil {
		c.Log.Item(err)
	}

	err = bson.Unmarshal(sout.Out, outPointer)
	if err != nil {
		return
	}

	c.connectionPool.Put(s)

	return
}

func (c *ServiceClient) getConnection(lvl int) (service ServiceResource, err error) {
	if lvl > 5 {
		err = errors.New("Unable to retrieve a valid connection to the service")
		return
	}

	conn, err := c.connectionPool.Get()

	if err != nil || c.isClosed(conn.(ServiceResource)) {
		if conn != nil {
			s := conn.(ServiceResource)

			s.closed = true
			c.connectionPool.Put(s)
		}

		return c.getConnection(lvl + 1)
	}

	service = conn.(ServiceResource)

	return service, err
}

func (c *ServiceClient) isClosed(service ServiceResource) bool {
	key := getInstanceKey(service.service)

	if _, ok := c.instances[key]; ok {
		return false
	}

	return true
}

func getInstanceKey(service service.Service) string {
	return service.Config.ServiceAddr.String()
}
