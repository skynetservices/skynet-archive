package skylib

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/bketelsen/skynet/skylib/util"
	"github.com/erikstmartin/msgpack-rpc/go/rpc"
	"math/rand"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type Client struct {
	DoozerConn DoozerConnection

	Config *ClientConfig
	Log    Logger `json:"-"`
}

type ServiceResource struct {
	conn    *rpc.Session
	service Service
	closed  bool
}

func (s ServiceResource) Close() {
	s.closed = true
	s.conn.Close()
}

func (s ServiceResource) IsClosed() bool {
	return s.closed
}

func (c *Client) doozer() DoozerConnection {
	if c.DoozerConn == nil {
		c.DoozerConn = NewDoozerConnectionFromConfig(*c.Config.DoozerConfig, c.Config.Log)

		c.DoozerConn.Connect()
	}

	return c.DoozerConn
}

func NewClient(config *ClientConfig) *Client {
	if config.Log == nil {
		config.Log = NewConsoleLogger(os.Stderr)
	}

	if config.ConnectionPoolSize == 0 {
		config.ConnectionPoolSize = 1
	}

	client := &Client{
		Config:     config,
		DoozerConn: NewDoozerConnectionFromConfig(*config.DoozerConfig, config.Log),
		Log:        config.Log,
	}

	client.Log.Item(config)

	client.DoozerConn.Connect()

	return client
}

// This will not fail if no services currently exist, this saves from chicken and egg issues with dependencies between services
// TODO: We should probably determine a way of supplying secondary conditions, for example it's ok to go to a different data center only if there are no instances in our current datacenter
func (c *Client) GetService(name string, version string, region string, host string) *ServiceClient {
	var conn net.Conn
	var err error

	registered := true

	service := &ServiceClient{
		Log:            c.Config.Log,
		connectionPool: pools.NewRoundRobin(c.Config.ConnectionPoolSize, c.Config.IdleTimeout),
		query: &Query{
			DoozerConn: c.DoozerConn,
			Service:    name,
			Version:    version,
			Host:       host,
			Region:     region,
			Registered: &registered,
		},
		instances: make(map[string]Service, 0),
	}

	// Load initial list of instances
	results := service.query.FindInstances()

	if results != nil {
		for _, instance := range *results {
			key := instance.Config.ServiceAddr.IPAddress + ":" + strconv.Itoa(instance.Config.ServiceAddr.Port)
			service.instances[key] = *instance
		}
	}

	go service.monitorInstances()

	factory := func() (pools.Resource, error) {
		if len(service.instances) < 1 {

			return nil, errors.New("No services available that match your criteria")
		}

		// Connect to random instance
		key := (rand.Int() % len(service.instances))

		var instance Service

		i := 0

		for _, v := range service.instances {
			if i == key {
				instance = v
				break
			}
		}

		conn, err = net.Dial("tcp", instance.Config.ServiceAddr.IPAddress+":"+strconv.Itoa(instance.Config.ServiceAddr.Port))

		if err != nil {
			// TODO: handle failure here and attempt to connect to a different instance
			return nil, errors.New("Failed to connect to service: " + instance.Config.ServiceAddr.IPAddress + ":" + strconv.Itoa(instance.Config.ServiceAddr.Port))
		}

		resource := ServiceResource{
			conn:    rpc.NewSession(conn, true),
			service: instance,
		}

		return resource, nil
	}

	service.connectionPool.Open(factory)

	return service
}

type ServiceClient struct {
	Log            Logger `json:"-"`
	connectionPool *pools.RoundRobin
	query          *Query
	instances      map[string]Service
}

func (c *ServiceClient) monitorInstances() {
	// TODO: Let's watch doozer and keep this list up to date so we don't need to search it every time we spawn a new connection
	doozer := c.query.DoozerConn

	rev := doozer.GetCurrentRevision()

	for {
		ev, err := doozer.Wait("/services/**", rev+1)
		rev = ev.Rev

		if err == nil {
			var service Service

			buf := bytes.NewBuffer(ev.Body)

			err = json.Unmarshal(buf.Bytes(), &service)

			if err == nil {
				parts := strings.Split(ev.Path, "/")

				if c.query.pathMatches(parts, ev.Path) {
					key := service.Config.ServiceAddr.IPAddress + ":" + strconv.Itoa(service.Config.ServiceAddr.Port)

					if service.Registered == true {
						c.Log.Println("New Service Instance Discovered: " + key)
						c.instances[key] = service
					} else {
						c.Log.Println("Service Instance Removed: " + key)
						delete(c.instances, key)
					}
				}
			}
		}
	}
}

func (c *ServiceClient) Send(funcName string, arguments ...interface{}) (reflect.Value, error) {
	// TODO: timeout logic
	service, _ := c.getConnection(0)

	// TODO: Check for connectivity issue so that we can try to get another resource out of the pool
	val, er := service.conn.SendV(funcName, arguments)

	c.connectionPool.Put(service)

	return val, er
}

func (c *ServiceClient) getConnection(lvl int) (ServiceResource, error) {
	if lvl > 5 {
		panic("Unable to retrieve a valid connection to the service")
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

	service := conn.(ServiceResource)

	return service, err
}

func (c *ServiceClient) isClosed(service ServiceResource) bool {
	key := getInstanceKey(service.service)

	if _, ok := c.instances[key]; ok {
		return false
	}

	return true
}

func getInstanceKey(service Service) string {
	return service.Config.ServiceAddr.IPAddress + ":" + strconv.Itoa(service.Config.ServiceAddr.Port)
}
