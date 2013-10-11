package client

import (
	"errors"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/client/loadbalancer/roundrobin"
	"github.com/skynetservices/skynet/test"
	"testing"
	"time"
)

// TODO: Validate MaxConnectionsToInstance and MaxIdleConnectionsToInstance
var serviceManager = new(test.ServiceManager)

func init() {
	skynet.SetServiceManager(serviceManager)
}

func TestGetNetwork(t *testing.T) {
	// Default to tcp
	if GetNetwork() != "tcp" {
		t.Fatal("GetNetwork() returned incorrect value")
	}
}

func TestSetNetwork(t *testing.T) {
	for _, n := range knownNetworks {
		err := SetNetwork(n)

		if err != nil {
			t.Fatal("SetNetwork() incorrectly rejected known network")
		}
	}

	err := SetNetwork("foo")

	if err == nil {
		t.Fatal("SetNetwork() accepted invalid network")
	}
}

func TestGetServiceFromCriteria(t *testing.T) {
	c := &skynet.Criteria{
		Services: []skynet.ServiceCriteria{
			skynet.ServiceCriteria{Name: "TestService"},
		},
	}

	c.AddRegion("Tampa")
	service := GetServiceFromCriteria(c)
	defer resetClient()

	if service.(*ServiceClient).criteria != c {
		t.Fatal("GetServiceFromCriteria() failed to associate critera with ServiceClient")
	}
}

func TestClientClose(t *testing.T) {
	closeCalled := false

	sc := test.ServiceClient{}
	sc.CloseFunc = func() {
		closeCalled = true
	}

	addServiceClient(ServiceClientProvider(&sc))
	defer resetClient()

	Close()

	if !closeCalled {
		t.Fatal("Close() is expected to call Close() on all ServiceClients")
	}
}

func TestInstanceNotificationForwardedToServiceClient(t *testing.T) {
	timeout := 5 * time.Millisecond

	watch := make(chan interface{})
	receive := make(chan interface{})

	// watching isn't started till we have at least one ServiceClient
	sc := test.ServiceClient{
		MatchesFunc: func(s skynet.ServiceInfo) bool {
			return true
		},
		NotifyFunc: func(s skynet.InstanceNotification) {
			watch <- true
		},
	}

	addServiceClient(ServiceClientProvider(&sc))
	defer resetClient()

	si := serviceInfo()

	// Add
	go receiveOrTimeout(watch, receive, timeout)
	go sendInstanceNotification(skynet.InstanceAdded, *si)

	v := <-receive
	if _, fail := v.(error); fail {
		t.Fatal("Notify() is expected to be called on all matching ServiceClients")
	}

	// Update
	si.Registered = false

	go receiveOrTimeout(watch, receive, timeout)
	go sendInstanceNotification(skynet.InstanceUpdated, *si)

	v = <-receive
	if _, fail := v.(error); fail {
		t.Fatal("Notify() is expected to be called on all matching ServiceClients")
	}

	// Remove
	go receiveOrTimeout(watch, receive, timeout)
	go sendInstanceNotification(skynet.InstanceRemoved, *si)

	v = <-receive
	if _, fail := v.(error); fail {
		t.Fatal("Notify() is expected to be called on all matching ServiceClients")
	}

}

func TestInstanceNotificationsUpdatePool(t *testing.T) {
	watch := make(chan interface{})
	receive := make(chan interface{})
	timeout := 5 * time.Millisecond

	// watching isn't started till we have at least one ServiceClient
	sc := test.ServiceClient{
		MatchesFunc: func(s skynet.ServiceInfo) bool {
			return true
		},
	}

	addServiceClient(ServiceClientProvider(&sc))
	defer resetClient()

	si := serviceInfo()

	// Add
	pool = &test.Pool{
		AddInstanceFunc: func(s skynet.ServiceInfo) {
			watch <- true
		},
	}

	go receiveOrTimeout(watch, receive, timeout)
	go sendInstanceNotification(skynet.InstanceAdded, *si)

	v := <-receive
	if _, fail := v.(error); fail {
		t.Fatal("Failed to notify Pool of InstanceNotification")
	}

	// Update
	si.Registered = false
	pool = &test.Pool{
		UpdateInstanceFunc: func(s skynet.ServiceInfo) {
			watch <- true
		},
	}

	go receiveOrTimeout(watch, receive, timeout)
	go sendInstanceNotification(skynet.InstanceUpdated, *si)

	v = <-receive
	if _, fail := v.(error); fail {
		t.Fatal("Failed to notify Pool of InstanceNotification")
	}

	// Remove
	pool = &test.Pool{
		RemoveInstanceFunc: func(s skynet.ServiceInfo) {
			watch <- true
		},
	}

	go receiveOrTimeout(watch, receive, timeout)
	go sendInstanceNotification(skynet.InstanceRemoved, *si)

	v = <-receive
	if _, fail := v.(error); fail {
		t.Fatal("Failed to notify Pool of InstanceNotification")
	}
}

func serviceInfo() *skynet.ServiceInfo {
	si := skynet.NewServiceInfo("TestService", "1.0.0")
	si.Registered = true

	return si
}

func resetClient() {
	serviceClients = []ServiceClientProvider{}

	network = "tcp"
	knownNetworks = []string{"tcp", "tcp4", "tcp6", "udp", "udp4", "udp6", "ip", "ip4", "ip6", "unix", "unixgram", "unixpacket"}

	pool = NewPool()
	LoadBalancerFactory = roundrobin.New
}

func sendInstanceNotification(typ int, si skynet.ServiceInfo) {
	instanceWatcher <- skynet.InstanceNotification{Type: typ, Service: si}
}

func receiveOrTimeout(watchChan chan interface{}, respChan chan interface{}, d time.Duration) {
	timeout := time.After(d)

	select {
	case v := <-watchChan:
		respChan <- v
		return
	case <-timeout:
		respChan <- errors.New("timeout")
		return
	}
}
