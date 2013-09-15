package client

import (
	"errors"
	"github.com/skynetservices/skynet2"
	"github.com/skynetservices/skynet2/client/conn"
	"github.com/skynetservices/skynet2/client/loadbalancer"
	"github.com/skynetservices/skynet2/client/loadbalancer/roundrobin"
	"github.com/skynetservices/skynet2/config"
	"github.com/skynetservices/skynet2/log"
	"sync"
	"time"
)

const (
	DIAL_TIMEOUT = 500 * time.Millisecond
)

func init() {
	go mux()
}

// TODO: implement a way to report/remove instances that fail a number of times
var (
	network        = "tcp"
	knownNetworks  = []string{"tcp", "tcp4", "tcp6", "udp", "udp4", "udp6", "ip", "ip4", "ip6", "unix", "unixgram", "unixpacket"}
	serviceClients = []ServiceClientProvider{}

	closeChan       = make(chan bool, 1)
	instanceWatcher = make(chan skynet.InstanceNotification, 100)

	pool                ConnectionPooler     = NewPool()
	LoadBalancerFactory loadbalancer.Factory = roundrobin.New
	waiter              sync.WaitGroup
)

var (
	UnknownNetworkError = errors.New("Unknown network")
)

/*
client.GetNetwork() returns the network used for client connections (default tcp)
tcp, tcp4, tcp6, udp, udp4, udp6, ip, ip4, ip6, unix, unixgram, unixpacket
*/
func GetNetwork() string {
	return network
}

/*
client.SetNetwork() sets the network used for client connections (default tcp)
tcp, tcp4, tcp6, udp, udp4, udp6, ip, ip4, ip6, unix, unixgram, unixpacket
*/
func SetNetwork(network string) error {
	for _, n := range knownNetworks {
		if n == network {
			return nil
		}
	}

	return UnknownNetworkError
}

/*
client.SetLoadBalancer() provide a custom load balancer to determine the order in which instances are sent requests
*/
func SetLoadBalancerFactory(factory loadbalancer.Factory) {
	LoadBalancerFactory = factory
}

/*
client.GetServiceFromCriteria() Returns a client specific to the skynet.Criteria provided.
Only instances that match this criteria will service the requests.

The core reason to use this over GetService() is that the load balancer will use the order of the criteria items to determine which datacenter it should roll over to first etc.
*/
func GetServiceFromCriteria(c *skynet.Criteria) ServiceClientProvider {
	sc := NewServiceClient(c)

	// This should block, we dont want to return the ServiceClient before the client has fully registered it
	addServiceClient(sc)

	return sc
}

/*
client.Close() Closes all ServiceClient's and releases their network resources
*/
func Close() {
	waiter.Add(1)

	closeChan <- true

	// Wait for all ServiceClient's to finish and close their connections
	waiter.Wait()
}

/*
client.GetService() Returns a client specific to the criteria provided
Empty values will be treated as wildcards and will be determined to match everything
*/
func GetService(name string, version string, region string, host string) ServiceClientProvider {
	criteria := &skynet.Criteria{
		Services: []skynet.ServiceCriteria{
			skynet.ServiceCriteria{Name: name, Version: version},
		},
	}

	if region != "" {
		criteria.AddRegion(region)
	}

	if host != "" {
		criteria.AddHost(host)
	}

	s := GetServiceFromCriteria(criteria)

	return s
}

func mux() {
	for {
		select {
		case n := <-instanceWatcher:
			updateInstance(n)
		case <-closeChan:
			for _, sc := range serviceClients {
				sc.Close()
			}

			pool.Close()
			serviceClients = []ServiceClientProvider{}
			waiter.Done()
		}
	}
}

/*
client.acquire will return an idle connection or a new one
*/
func acquire(s skynet.ServiceInfo) (c conn.Connection, err error) {
	return pool.Acquire(s)
}

/*
client.release will release a resource for use by others. If the idle queue is
full, the resource will be closed.
*/
func release(c conn.Connection) {
	pool.Release(c)
}

func addServiceClient(sc ServiceClientProvider) {
	serviceClients = append(serviceClients, sc)

	instances := skynet.GetServiceManager().Watch(sc, instanceWatcher)

	for _, i := range instances {
		pool.AddInstance(i)
		sc.Notify(skynet.InstanceNotification{Type: skynet.InstanceAdded, Service: i})
	}
}

// TODO: we need a method here to removeServiceClient
// it should remove instances from the pool if there are no remaining
// ServiceClients that match the instance, and should end the watch from ServiceManager

// only call from mux()
func updateInstance(n skynet.InstanceNotification) {
	// Forward notification on to ServiceClients that match
	for _, sc := range serviceClients {
		if sc.Matches(n.Service) {
			go sc.Notify(n)
		}
	}

	// Update our internal pools
	switch n.Type {
	case skynet.InstanceAdded:
		go pool.AddInstance(n.Service)
	case skynet.InstanceUpdated:
		go pool.UpdateInstance(n.Service)
	case skynet.InstanceRemoved:
		go pool.RemoveInstance(n.Service)
	}

}

func getIdleConnectionsToInstance(s skynet.ServiceInfo) int {
	if n, err := config.Int(s.Name, s.Version, "client.conn.idle"); err == nil {
		return n
	}

	return config.DefaultIdleConnectionsToInstance
}

func getMaxConnectionsToInstance(s skynet.ServiceInfo) int {
	if n, err := config.Int(s.Name, s.Version, "client.conn.max"); err == nil {
		return n
	}

	return config.DefaultMaxConnectionsToInstance
}

func getIdleTimeout(s skynet.ServiceInfo) time.Duration {
	if d, err := config.String(s.Name, s.Version, "client.timeout.idle"); err == nil {
		if timeout, err := time.ParseDuration(d); err != nil {
			return timeout
		}

		log.Println(log.ERROR, "Failed to parse client.timeout.idle", err)
	}

	return config.DefaultIdleTimeout
}
