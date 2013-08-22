package client

import (
	"errors"
	"github.com/skynetservices/skynet2"
	"github.com/skynetservices/skynet2/client/loadbalancer"
	"github.com/skynetservices/skynet2/log"
	"reflect"
	"sync"
	"time"
)

// TODO: Implement SendTimeout()
// TODO: Implement SendOnceTimeout()

var (
	ServiceClientClosed = errors.New("Service client shutdown")
	RequestTimeout      = errors.New("Request timed out")
)

/*
ServiceSender Responsible for sending requests to the cluster.
This is mostly used as way to test that clients make appropriate requests to services without the need to run those services
*/
type ServiceClientProvider interface {
	SetDefaultTimeout(retry, giveup time.Duration)
	GetDefaultTimeout() (retry, giveup time.Duration)

	Close()

	Send(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) (err error)
	SendOnce(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) (err error)

	Notify(n skynet.InstanceNotification)
	Matches(n skynet.ServiceInfo) bool
}

type ServiceClient struct {
	loadBalancer loadbalancer.LoadBalancer
	criteria     *skynet.Criteria
	shutdown     bool
	closed       bool

	retryTimeout  time.Duration
	giveupTimeout time.Duration

	waiter sync.WaitGroup

	// mux channels
	muxChan               chan interface{}
	instanceNotifications chan skynet.InstanceNotification
	timeoutChan           chan timeoutLengths
	shutdownChan          chan bool

	// TODO: remove this if we dont need it, but i think we need it for items that go into the RequestInfo
	//cconfig       skynet.ClientConfig
}

/*
client.NewServiceClient Initializes a new ClientService
*/
func NewServiceClient(c *skynet.Criteria) ServiceClientProvider {
	sc := &ServiceClient{
		criteria:              c,
		instanceNotifications: make(chan skynet.InstanceNotification, 100),
		timeoutChan:           make(chan timeoutLengths),
		shutdownChan:          make(chan bool),
		muxChan:               make(chan interface{}),
		loadBalancer:          LoadBalancerFactory([]skynet.ServiceInfo{}),

		retryTimeout:  skynet.DefaultRetryDuration,
		giveupTimeout: skynet.DefaultTimeoutDuration,
	}

	go sc.mux()

	return sc
}

/*
ServiceClient.Send() will send a request to one of the available instances. In intervals of retry time,
it will send additional requests to other known instances. If no response is heard after
the giveup time has passed, it will return an error.
*/
func (c *ServiceClient) Send(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) (err error) {
	if c.closed {
		return ServiceClientClosed
	}

	retry, giveup := c.GetDefaultTimeout()
	return c.send(retry, giveup, ri, fn, in, out)
}

/*
ServiceClient.SendOnce() will send a request to one of the available instances. If no response is heard after
the giveup time has passed, it will return an error.
*/
func (c *ServiceClient) SendOnce(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) (err error) {
	if c.closed {
		return ServiceClientClosed
	}
	_, giveup := c.GetDefaultTimeout()
	return c.send(0, giveup, ri, fn, in, out)
}

/*
ServiceClient.SetTimeout() sets the time before ServiceClient.Send() retries requests, and
the time before ServiceClient.Send() and ServiceClient.SendOnce() give up. Setting retry
or giveup to 0 indicates no retry or time out.
*/
func (c *ServiceClient) SetDefaultTimeout(retry, giveup time.Duration) {
	c.muxChan <- timeoutLengths{
		retry:  retry,
		giveup: giveup,
	}
}

/*
ServiceClient.GetTimeout() returns current timeout values
*/
func (c *ServiceClient) GetDefaultTimeout() (retry, giveup time.Duration) {
	tls := <-c.timeoutChan
	retry, giveup = tls.retry, tls.giveup

	return
}

/*
ServiceClient.Close() refuses any new requests, and waits for active requests to finish
*/
func (c *ServiceClient) Close() {
	c.shutdownChan <- true
	c.waiter.Wait()
}

/*
ServiceClient.NewRequestInfo() create a new RequestInfo object specific to this service
*/
func (c *ServiceClient) NewRequestInfo() (ri *skynet.RequestInfo) {
	// TODO: Set
	ri = &skynet.RequestInfo{
		RequestID: skynet.UUID(),
	}

	return
}

/*
ServiceClient.Matches() determins if the provided Service matches the criteria associated with this client
*/
func (c *ServiceClient) Matches(s skynet.ServiceInfo) bool {
	return c.criteria.Matches(s)
}

/*
ServiceClient.Notify() Update available instances based off provided InstanceNotification
*/
func (c *ServiceClient) Notify(n skynet.InstanceNotification) {
	c.instanceNotifications <- n
}

func (c *ServiceClient) send(retry, giveup time.Duration, ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) (err error) {
	if ri == nil {
		ri = c.NewRequestInfo()
	}

	attempts := make(chan sendAttempt)

	retryChan := make(chan bool, 1)

	var retryTicker <-chan time.Time
	if retry > 0 {
		retryTicker = time.Tick(retry)
	}

	var timeoutTimer <-chan time.Time
	if giveup > 0 {
		timeoutTimer = time.NewTimer(giveup).C
	}

	attemptCount := 1
	go c.attemptSend(retry, attempts, ri, fn, in, out)

	for {
		select {
		case <-retryTicker:
			retryChan <- true
		case <-retryChan:
			attemptCount++
			ri.RetryCount++
			go c.attemptSend(retry, attempts, ri, fn, in, out)

		case <-timeoutTimer:
			err = RequestTimeout
			return

		case attempt := <-attempts:
			if attempt.err != nil {
				log.Println(log.ERROR, "Attempt Error: ", attempt.err)

				// If there is no retry timer we should retry after each failed attempt
				if retryTicker == nil {
					retryChan <- true
				}

				continue
			}

			// copy into the caller's value
			v := reflect.Indirect(reflect.ValueOf(out))
			v.Set(reflect.Indirect(reflect.ValueOf(attempt.result)))

			return
		}
	}
}

type sendAttempt struct {
	err    error
	result interface{}
}

func (c *ServiceClient) attemptSend(timeout time.Duration, attempts chan sendAttempt, ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) {
	s, err := c.loadBalancer.Choose()

	if err != nil {
		attempts <- sendAttempt{err: err}
		return
	}

	conn, err := acquire(s)
	defer release(conn)

	if err != nil {
		attempts <- sendAttempt{err: err}
		return
	}

	// Create a new instance of the type, we dont want race conditions where 2 connections are unmarshalling to the same object
	res := sendAttempt{
		result: reflect.New(reflect.Indirect(reflect.ValueOf(out)).Type()).Interface(),
	}

	err = conn.SendTimeout(ri, fn, in, res.result, timeout)

	if err != nil {
		res.err = err
	}

	attempts <- res
}

type timeoutLengths struct {
	retry, giveup time.Duration
}

func (c *ServiceClient) mux() {
	for {
		select {
		case mi := <-c.muxChan:
			switch m := mi.(type) {
			case timeoutLengths:
				c.retryTimeout = m.retry
				c.giveupTimeout = m.giveup
			}
		case n := <-c.instanceNotifications:
			c.handleInstanceNotification(n)

		case c.timeoutChan <- timeoutLengths{
			retry:  c.retryTimeout,
			giveup: c.giveupTimeout,
		}:

		case shutdown := <-c.shutdownChan:
			// TODO: Close out all channels, and this goroutine after waiting for requests to finish
			if shutdown {
				c.closed = true
				return
			}
		}
	}
}

// this should only be called by mux()
func (c *ServiceClient) handleInstanceNotification(n skynet.InstanceNotification) {
	// TODO: ensure LoadBalancer is thread safe and call these as goroutines
	switch n.Type {
	case skynet.InstanceAdded:
		c.loadBalancer.AddInstance(n.Service)
	case skynet.InstanceUpdated:
		c.loadBalancer.UpdateInstance(n.Service)
	case skynet.InstanceRemoved:
		c.loadBalancer.RemoveInstance(n.Service)
	}
}
