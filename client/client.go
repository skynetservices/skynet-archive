package client

import (
	"errors"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/service"
	"net/rpc"
	"os"
)

var (
	ErrServiceUnregistered = errors.New("Service is unregistered")
	ErrRequestTimeout      = errors.New("Service request timed out")
)

type ServiceResource struct {
	rpcClient *rpc.Client
	service   *service.Service
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
	DoozerConn skynet.DoozerConnection

	Config *skynet.ClientConfig
	Log    skynet.Logger `json:"-"`
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

	s = newServiceClient(q, c)

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

func getInstanceKey(service *service.Service) string {
	return service.Config.ServiceAddr.String()
}
