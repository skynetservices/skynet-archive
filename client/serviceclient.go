package client

import (
	"errors"
	"fmt"
	"github.com/skynetservices/skynet2"
	"github.com/skynetservices/skynet2/log"
	"github.com/skynetservices/skynet2/pools"
	"labix.org/v2/mgo/bson"
	"time"
)

type serviceError struct {
	msg string
}

func (se serviceError) Error() string {
	return se.msg
}

type ServiceClientInterface interface {
	SetTimeout(retry, giveup time.Duration)
	GetTimeout() (retry, giveup time.Duration)
	Send(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) (err error)
	SendOnce(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) (err error)
}

type ServiceClient struct {
	client   *Client
	cconfig  *skynet.ClientConfig
	criteria *skynet.Criteria
	// a list of the known instances
	instances map[string]*servicePool

	muxChan     chan interface{}
	timeoutChan chan timeoutLengths

	listenID string

	retryTimeout  time.Duration
	giveupTimeout time.Duration

	servicePool chan *servicePool
	updateChan  <-chan time.Time
}

func (c *ServiceClient) Close() {
	for _, sp := range c.instances {
		sp.pool.Close()
	}
}

func newServiceClient(criteria *skynet.Criteria, c *Client) (sc *ServiceClient) {
	sc = &ServiceClient{
		client:        c,
		cconfig:       c.Config,
		criteria:      criteria,
		instances:     make(map[string]*servicePool),
		muxChan:       make(chan interface{}),
		timeoutChan:   make(chan timeoutLengths),
		retryTimeout:  skynet.DefaultRetryDuration,
		giveupTimeout: skynet.DefaultTimeoutDuration,
		servicePool:   make(chan *servicePool, 100),
		updateChan:    time.Tick(15 * time.Second),
	}
	sc.listenID = skynet.UUID()

	go sc.mux()

	instances, err := skynet.GetServiceManager().ListInstances(sc.criteria)

	if err == nil && len(instances) > 0 {
		for _, instance := range instances {
			sc.addInstanceMux(instance)
		}
	}

	go sc.managePools()

	return
}

type servicePool struct {
	service *skynet.ServiceInfo
	pool    *pools.ResourcePool
}

type timeoutLengths struct {
	retry, giveup time.Duration
}

func (c *ServiceClient) addInstanceMux(instance skynet.ServiceInfo) {
	m := skynet.ServiceDiscovered{&instance}
	key := getInstanceKey(&instance)
	_, known := c.instances[key]
	if !known {
		// we got a new pool, put it into the wild
		pool := c.client.getServicePool(m.Service)
		c.instances[key] = pool
	}
}

func (c *ServiceClient) removeInstanceMux(instance skynet.ServiceInfo) {
	key := getInstanceKey(&instance)
	_, known := c.instances[key]
	if !known {
		return
	}
	delete(c.instances, key)
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
		case c.timeoutChan <- timeoutLengths{
			retry:  c.retryTimeout,
			giveup: c.giveupTimeout,
		}:

		}
	}
}

// TODO: This is a short term solution to keeping the pools up to date with zookeeper, and load balancing across them
// to be replaced by full implementation later, with proper load balancing based off host metrics, and region/host priorities
func (c *ServiceClient) managePools() {
	for {
		for _, p := range c.instances {
			select {
			case <-c.updateChan:
				var currentInstances = make(map[string]*skynet.ServiceInfo)

				instances, err := skynet.GetServiceManager().ListInstances(c.criteria)
				if err == nil && len(instances) > 0 {
					for _, instance := range instances {
						key := getInstanceKey(&instance)
						currentInstances[key] = &instance
						c.addInstanceMux(instance)
					}
				}

				// Remove old instances
				for key, _ := range c.instances {
					if i, ok := currentInstances[key]; !ok {
						c.removeInstanceMux(*i)
					}
				}

				break
			case c.servicePool <- p:
			}
		}
	}
}

/*
ServiceClient.SetTimeout() sets the time before ServiceClient.Send() retries requests, and
the time before ServiceClient.Send() and ServiceClient.SendOnce() give up. Setting retry
or giveup to 0 indicates no retry or time out.
*/
func (c *ServiceClient) SetTimeout(retry, giveup time.Duration) {
	c.muxChan <- timeoutLengths{
		retry:  retry,
		giveup: giveup,
	}
}

func (c *ServiceClient) GetTimeout() (retry, giveup time.Duration) {
	tls := <-c.timeoutChan
	retry, giveup = tls.retry, tls.giveup
	return
}

/*
ServiceClient.Send() will send a request to one of the available instances. In intervals of retry time,
it will send additional requests to other known instances. If no response is heard after
the giveup time has passed, it will return an error.
*/
func (c *ServiceClient) Send(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) (err error) {
	retry, giveup := c.GetTimeout()
	return c.send(retry, giveup, ri, fn, in, out)
}

/*
ServiceClient.SendOnce() will send a request to one of the available instances. If no response is heard after
the giveup time has passed, it will return an error.
*/
func (c *ServiceClient) SendOnce(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) (err error) {
	_, giveup := c.GetTimeout()
	return c.send(0, giveup, ri, fn, in, out)
}

func (c *ServiceClient) send(retry, giveup time.Duration, ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) (err error) {
	if ri == nil {
		ri = &skynet.RequestInfo{
			RequestID: skynet.UUID(),
		}
	}

	attempts := make(chan sendAttempt)

	var ticker <-chan time.Time
	if retry > 0 {
		ticker = time.NewTicker(retry).C
	}

	var timeout <-chan time.Time
	if giveup > 0 {
		timeout = time.NewTimer(giveup).C
	}

	doneSignal := make(chan bool)
	attemptCount := 1

	defer func() {
		go func() {
			for i := 0; i < attemptCount; i++ {
				doneSignal <- true
			}
		}()
	}()

	go c.attemptSend(doneSignal, attempts, ri, fn, in)

	for {
		select {
		case <-ticker:
			attemptCount++
			ri.RetryCount++

			go c.attemptSend(doneSignal, attempts, ri, fn, in)
		case <-timeout:
			if err == nil {
				err = ErrRequestTimeout
			}
			// otherwise use the last error reported from an attempt
			return
		case attempt := <-attempts:
			err = attempt.err
			if err != nil {
				if _, ok := err.(serviceError); !ok {
					// error during transmition, abort this attempt
					if giveup == 0 {
						return
					}
					continue
				}
			}

			unmarshallerr := bson.Unmarshal(attempt.result, out)
			if unmarshallerr != nil {
				err = unmarshallerr
			}
			return
		}
	}

	return
}

type sendAttempt struct {
	result []byte
	err    error
}

func (c *ServiceClient) attemptSend(timeout chan bool,
	attempts chan sendAttempt, ri *skynet.RequestInfo,
	fn string, in interface{}) {

	// first find an available instance
	var r pools.Resource
	var err error
	for r == nil {
		if len(c.instances) < 1 {
			attempts <- sendAttempt{err: errors.New("No instances found")}
			return
		}

		sp := <-c.servicePool

		log.Println(log.TRACE, "Sending request to: "+sp.service.UUID)

		// then, get a connection to that instance
		r, err = sp.pool.Acquire()
		defer sp.pool.Release(r)
		if err != nil {
			if r != nil {
				r.Close()
			}
			// TODO: report connection failure
			failed := FailedConnection{err}
			log.Printf(log.ERROR, "%T: %+v", failed, failed)
		}
	}

	if err != nil {
		log.Printf(log.ERROR, "Error: %v", err)

		attempts <- sendAttempt{err: err}
		return
	}

	sr := r.(ServiceResource)

	result, serviceErr, err := c.sendToInstance(sr, ri, fn, in)
	if err != nil {
		// some communication error happened, shut down this connection and remove it from the pool
		sr.Close()
		return
	}

	attempts <- sendAttempt{
		result: result,
		err:    serviceErr,
	}
}

// ServiceClient.sendToInstance() tries to make an RPC request on a particular connection to an instance
func (c *ServiceClient) sendToInstance(sr ServiceResource,
	requestInfo *skynet.RequestInfo, funcName string, in interface{}) (
	result []byte, serviceErr, err error) {

	sin := skynet.ServiceRPCIn{
		RequestInfo: requestInfo,
		Method:      funcName,
		ClientID:    sr.clientID,
	}

	sin.In, err = bson.Marshal(in)
	if err != nil {
		err = fmt.Errorf("Error calling bson.Marshal: %v", err)
		return
	}

	sout := skynet.ServiceRPCOut{}

	err = sr.rpcClient.Call(sr.service.Name+".Forward", sin, &sout)
	if err != nil {
		sr.Close()

		// Log failure
		log.Printf(log.ERROR, "Error calling sr.rpcClient.Call: "+err.Error())
	}

	if sout.ErrString != "" {
		serviceErr = serviceError{sout.ErrString}
	}

	result = sout.Out

	return
}
