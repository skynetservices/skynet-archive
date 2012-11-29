package client

import (
	"fmt"
	"github.com/4ad/doozer"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/pools"
	"labix.org/v2/mgo/bson"
	"time"
)

const DEBUG = false

func dbg(items ...interface{}) {
	if DEBUG {
		fmt.Println(items...)
	}
}

func dbgf(format string, items ...interface{}) {
	if DEBUG {
		fmt.Printf(format, items...)
	}
}

func dbgerr(name string, err error) {
	if err != nil {
		dbgf("(%s) %v\n", name, err)
	}
}

const TRACE = false

func ts(name string, items ...interface{}) {
	if TRACE {
		fmt.Printf("+%s %v\n", name, items)
	}
}
func te(name string, items ...interface{}) {
	if TRACE {
		fmt.Printf("-%s %v\n", name, items)
	}
}

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
	client  *Client
	Log     skynet.SemanticLogger `json:"-"`
	cconfig *skynet.ClientConfig
	query   *skynet.Query
	// a list of the known instances
	instances map[string]*servicePool

	chooser *InstanceChooser

	muxChan     chan interface{}
	timeoutChan chan timeoutLengths

	instanceListener *InstanceListener
	listenID         string

	retryTimeout  time.Duration
	giveupTimeout time.Duration
}

func newServiceClient(query *skynet.Query, c *Client) (sc *ServiceClient) {
	sc = &ServiceClient{
		client:        c,
		Log:           c.Config.Log,
		cconfig:       c.Config,
		query:         query,
		instances:     make(map[string]*servicePool),
		chooser:       NewInstanceChooser(c),
		muxChan:       make(chan interface{}),
		timeoutChan:   make(chan timeoutLengths),
		retryTimeout:  skynet.DefaultRetryDuration,
		giveupTimeout: skynet.DefaultTimeoutDuration,
	}
	sc.listenID = skynet.UUID()
	sc.instanceListener = c.instanceMonitor.Listen(sc.listenID, query, true)

	go sc.mux()
	return
}

type instanceFileCollector struct {
	files []string
}

func (ic *instanceFileCollector) VisitDir(path string, f *doozer.FileInfo) bool {
	return true
}
func (ic *instanceFileCollector) VisitFile(path string, f *doozer.FileInfo) {
	ic.files = append(ic.files, path)
}

type servicePool struct {
	service *skynet.ServiceInfo
	pool    *pools.ResourcePool
}

type timeoutLengths struct {
	retry, giveup time.Duration
}

func (c *ServiceClient) addInstanceMux(instance *skynet.ServiceInfo) {
	m := skynet.ServiceDiscovered{instance}
	key := getInstanceKey(m.Service)
	_, known := c.instances[key]
	if !known {
		// we got a new pool, put it into the wild
		c.instances[key] = c.client.getServicePool(m.Service)
		c.chooser.Add(m.Service)
		// Log event
		c.Log.Debug(fmt.Sprintf("%T: %+v", m, m))
	}
}

func (c *ServiceClient) removeInstanceMux(instance *skynet.ServiceInfo) {
	m := skynet.ServiceRemoved{instance}
	key := m.Service.Config.ServiceAddr.String()
	_, known := c.instances[key]
	if !known {
		return
	}
	c.chooser.Remove(m.Service)
	delete(c.instances, m.Service.Config.ServiceAddr.String())
	// Log event
	c.Log.Trace(fmt.Sprintf("%T: %+v", m, m))
}

func (c *ServiceClient) mux() {

	for {
		select {
		case ns := <-c.instanceListener.NotificationChan:
			for _, n := range ns {
				switch n.Type {
				case InstanceAddNotification, InstanceUpdateNotification:
					if n.Service.Registered {
						c.addInstanceMux(&n.Service)
					} else {
						c.removeInstanceMux(&n.Service)
					}
				case InstanceRemoveNotification:
					c.removeInstanceMux(&n.Service)
				}
			}
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

	ts("attemptSend")
	defer te("attemptSend")

	// first find an available instance
	var instance *skynet.ServiceInfo
	var r pools.Resource
	var err error
	for r == nil {
		var ok bool
		instance, ok = c.chooser.Choose(timeout)
		if !ok {
			dbg("timed out")
			// must have timed out
			return
		}
		dbg("chose", getInstanceKey(instance))
		sp := c.instances[getInstanceKey(instance)]

		// then, get a connection to that instance
		dbg("acquiring connection")
		r, err = sp.pool.Acquire()
		dbgerr("sp.pool.Acquire", err)
		dbg("acquired connection")
		defer sp.pool.Release(r)
		if err != nil {
			if r != nil {
				r.Close()
			}
			// TODO: report connection failure
			c.chooser.Remove(instance)
			// Log failure
			failed := FailedConnection{err}
			c.Log.Error(fmt.Sprintf("%T: %+v", failed, failed))
		}
	}

	if err != nil {
		c.Log.Error(fmt.Sprintf("Error: %v", err))

		attempts <- sendAttempt{err: err}
		return
	}

	sr := r.(ServiceResource)

	result, serviceErr, err := c.sendToInstance(sr, ri, fn, in)
	dbgerr("c.sendToInstance", err)
	if err != nil {
		// some communication error happened, shut down this connection and remove it from the pool
		sr.Close()
		// and remove the instance from the chooser
		c.chooser.Remove(instance)
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
	ts("sendToInstance", requestInfo)
	defer te("sendToInstance", requestInfo)

	sr.service.FetchStats(c.client.doozer())
	dbgf("stats: %+v\n", sr.service.Stats)

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

	err = sr.rpcClient.Call(sr.service.Config.Name+".Forward", sin, &sout)
	if err != nil {
		sr.Close()
		dbg("(sr.rpcClient.Call)", err)

		// Log failure
		c.Log.Error("Error calling sr.rpcClient.Call: " + err.Error())
	}

	if sout.ErrString != "" {
		serviceErr = serviceError{sout.ErrString}
	}

	result = sout.Out

	return
}
