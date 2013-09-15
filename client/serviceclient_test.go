package client

import (
	"github.com/skynetservices/skynet2"
	"github.com/skynetservices/skynet2/client/conn"
	"github.com/skynetservices/skynet2/test"
	"labix.org/v2/mgo/bson"
	"testing"
	"time"
)

// TODO: Test Send()
// TODO: Test SendOnce()
// TODO: Test SendTimeout()
// TODO: Test SendOnceTimeout()
// TODO: Test Timeout/Retry logic
// TODO: Test Close() closes connections
// TODO: Test that NewServiceClient gets a LoadBalancer from the factory
// TODO: Test error conditions from connection

func TestSetTimeout(t *testing.T) {
	defer resetClient()

	s := GetService("foo", "1.0.0", "", "")

	s.SetDefaultTimeout(5*time.Second, 10*time.Second)

	retry, giveup := s.GetDefaultTimeout()

	if retry != 5*time.Second || giveup != 10*time.Second {
		t.Fatal("SetDefaultTimeout() timeout values not set correctly")
	}
}

func TestInstanceNotificationsUpdateLoadBalancer(t *testing.T) {
	defer resetClient()

	watch := make(chan interface{})
	receive := make(chan interface{})
	timeout := 5 * time.Millisecond

	// watching isn't started till we have at least one ServiceClient
	criteria := &skynet.Criteria{}
	sc := NewServiceClient(criteria)
	sClient := sc.(*ServiceClient)

	// Use stub pool, we dont want real connections being made
	pool = &test.Pool{}

	addServiceClient(sc)

	si := serviceInfo()
	si.UUID = "FOO"

	// Add
	sClient.loadBalancer = &test.LoadBalancer{
		AddInstanceFunc: func(s skynet.ServiceInfo) {
			watch <- true
		},
	}

	go receiveOrTimeout(watch, receive, timeout)
	go sendInstanceNotification(skynet.InstanceAdded, *si)

	v := <-receive
	if _, fail := v.(error); fail {
		t.Fatal("Failed to notify LoadBalancer of InstanceNotification")
	}

	// Update
	si.Registered = false
	sClient.loadBalancer = &test.LoadBalancer{
		UpdateInstanceFunc: func(s skynet.ServiceInfo) {
			watch <- true
		},
	}

	go receiveOrTimeout(watch, receive, timeout)
	go sendInstanceNotification(skynet.InstanceUpdated, *si)

	v = <-receive
	if _, fail := v.(error); fail {
		t.Fatal("Failed to notify LoadBalancer of InstanceNotification")
	}

	// Remove
	sClient.loadBalancer = &test.LoadBalancer{
		RemoveInstanceFunc: func(s skynet.ServiceInfo) {
			watch <- true
		},
	}

	go receiveOrTimeout(watch, receive, timeout)
	go sendInstanceNotification(skynet.InstanceRemoved, *si)

	v = <-receive
	if _, fail := v.(error); fail {
		t.Fatal("Failed to notify LoadBalancer of InstanceNotification")
	}
}

func TestCloseRefusesNewRequests(t *testing.T) {
	s := GetService("foo", "1.0.0", "", "")
	s.Close()

	var val string

	err := s.Send(nil, "Foo", val, &val)
	if err != ServiceClientClosed {
		t.Fatal("Close() did not refuse new requests")
	}

	err = s.SendOnce(nil, "Foo", val, &val)
	if err != ServiceClientClosed {
		t.Fatal("Close() did not refuse new requests")
	}
}

func TestSend(t *testing.T) {
	called := false

	type r struct {
		Bar string
	}

	request := 20
	response := r{""}

	sc := GetService("foo", "1.0.0", "", "")
	stubForSend(sc, func(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) (err error) {
		called = true

		resp := r{
			"Foo",
		}

		b, err := bson.Marshal(resp)

		if err != nil {
			t.Error(err)
		}

		err = bson.Unmarshal(b, out)

		return err
	})

	err := sc.Send(nil, "bar", request, &response)

	if err != nil {
		t.Error(err)
	}

	if !called {
		t.Fatal("Send not called")
	}

	if response.Bar != "Foo" {
		t.Fatal("response value failed to copy")
	}
}

// Helper for validating and testing send logic
// stubs ServiceManager, Pool, Connection, LoadBalancer
func stubForSend(sc ServiceClientProvider, f func(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) (err error)) {
	sm := &test.ServiceManager{}
	skynet.SetServiceManager(skynet.ServiceManager(sm))

	pool = &test.Pool{
		AcquireFunc: func(s skynet.ServiceInfo) (conn.Connection, error) {
			c := &test.Connection{
				SendTimeoutFunc: func(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}, timeout time.Duration) (err error) {
					return f(ri, fn, in, out)
				},
			}

			return c, nil
		},
	}

	sClient := sc.(*ServiceClient)
	sClient.loadBalancer = &test.LoadBalancer{
		ChooseFunc: func() (s skynet.ServiceInfo, err error) {
			return
		},
	}
}
