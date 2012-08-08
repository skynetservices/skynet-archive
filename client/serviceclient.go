package client

import (
	"bytes"
	"encoding/json"
	"github.com/4ad/doozer"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/pools"
	"github.com/bketelsen/skynet/service"
	"launchpad.net/mgo/v2/bson"
	"path"
	"reflect"
	"strings"
	"time"
)

type ServiceClient struct {
	client  *Client
	Log     skynet.Logger `json:"-"`
	cconfig *skynet.ClientConfig
	query   *Query
	// a list of the known instances
	instances map[string]*servicePool
	// a pool of the available instances. contains things of type servicePool
	instancePool *pools.ResourcePool
	muxChan      chan interface{}
	timeoutChan  chan timeoutLengths

	retryTimeout  time.Duration
	giveupTimeout time.Duration
}

func newServiceClient(query *Query, c *Client) (sc *ServiceClient) {
	sc = &ServiceClient{
		client:       c,
		Log:          c.Config.Log,
		cconfig:      c.Config,
		query:        query,
		instances:    make(map[string]*servicePool),
		instancePool: pools.NewResourcePool(func() (pools.Resource, error) { panic("unreachable") }, -1, 0),
		muxChan:      make(chan interface{}),
		timeoutChan:  make(chan timeoutLengths),
	}
	go sc.mux()
	go sc.monitorInstances()
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

func (c *ServiceClient) monitorInstances() {
	// TODO: Let's watch doozer and keep this list up to date so we don't need to search it every time we spawn a new connection
	doozer := c.query.DoozerConn

	rev := doozer.GetCurrentRevision()

	ddir := c.query.makePath()

	var ifc instanceFileCollector
	errch := make(chan error)
	doozer.Walk(rev, ddir, &ifc, errch)
	select {
	case err := <-errch:
		c.Log.Item(err)
	default:
	}

	for _, file := range ifc.files {
		buf, _, err := doozer.Get(file, rev)
		if err != nil {
			c.Log.Item(err)
			continue
		}
		var s service.Service
		err = json.Unmarshal(buf, &s)
		if err != nil {
			c.Log.Item(err)
			continue
		}

		c.muxChan <- service.ServiceDiscovered{
			Service: &s,
		}
	}

	watchPath := path.Join(c.query.makePath(), "**")

	for {
		ev, err := doozer.Wait(watchPath, rev+1)
		rev = ev.Rev
		if err != nil {
			continue
		}

		var s service.Service

		buf := bytes.NewBuffer(ev.Body)

		err = json.Unmarshal(buf.Bytes(), &s)
		if err != nil {
			continue
		}

		parts := strings.Split(ev.Path, "/")

		if c.query.pathMatches(parts, ev.Path) {
			//key := s.Config.ServiceAddr.String()

			if s.Registered == true {
				c.muxChan <- service.ServiceDiscovered{
					Service: &s,
				}
			} else {
				c.muxChan <- service.ServiceRemoved{
					Service: &s,
				}
			}
		}
	}
}

type servicePool struct {
	service *service.Service
	pool    *pools.ResourcePool
	closed  bool
}

// this is here to make it a pools.Resource
func (sp *servicePool) Close() {
	sp.closed = true
}

// this is here to make it a pools.Resource
func (sp *servicePool) IsClosed() bool {
	return sp.closed
}

type timeoutLengths struct {
	retry, giveup time.Duration
}

func (c *ServiceClient) mux() {

	for {
		select {
		case mi := <-c.muxChan:
			switch m := mi.(type) {
			case service.ServiceDiscovered:
				key := m.Service.Config.ServiceAddr.String()
				_, known := c.instances[key]
				if !known {
					// we got a new pool, put it into the wild
					c.instances[key] = c.client.getServicePool(m.Service)
					c.instancePool.Release(c.instances[key])
					c.Log.Item(m)
				}

			case service.ServiceRemoved:
				key := m.Service.Config.ServiceAddr.String()
				c.instances[key].Close()
				delete(c.instances, m.Service.Config.ServiceAddr.String())
				c.Log.Item(m)
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

// ServiceClient.sendToInstance() tries to make an RPC request on a particular connection to an instance
func (c *ServiceClient) sendToInstance(sr ServiceResource, requestInfo *skynet.RequestInfo, funcName string, in interface{}, outPointer interface{}) (err error) {
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
	err = sr.rpcClient.Call(sr.service.Config.Name+".Forward", sin, &sout)
	if err != nil {
		sr.Close()
		c.Log.Item(err)
	}

	err = bson.Unmarshal(sout.Out, outPointer)
	if err != nil {
		return
	}

	return
}

func cloneOutDest(outDest interface{}) (clone interface{}) {
	outType := reflect.TypeOf(outDest)
	switch outType.Kind() {
	case reflect.Ptr:
		clonePtr := reflect.New(outType.Elem())
		clone = clonePtr.Interface()
	case reflect.Map:
		cloneMap := reflect.MakeMap(outType)
		clone = cloneMap.Interface()
	default:
		panic("illegal out type")
	}
	return
}

func copyOutDest(outDest interface{}, src interface{}) {
	outType := reflect.TypeOf(outDest)
	outVal := reflect.ValueOf(outDest)
	srcVal := reflect.ValueOf(src)
	switch outType.Kind() {
	case reflect.Ptr:
		outVal.Elem().Set(srcVal.Elem())
	case reflect.Map:
		for _, key := range srcVal.MapKeys() {
			val := srcVal.MapIndex(key)
			outVal.SetMapIndex(key, val)
		}
	default:
		panic("illegal out type")
	}

}

func (c *ServiceClient) trySend(attempts chan sendAttempt, ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) {
	spr, _ := c.instancePool.Acquire()
	sp := spr.(*servicePool)
	defer c.instancePool.Release(sp)

	// then, get a connection to that instance
	r, err := sp.pool.Acquire()
	defer sp.pool.Release(r)
	if err != nil {
		c.Log.Item(err)
		attempts <- sendAttempt{err: err}
		return
	}

	sr := r.(ServiceResource)

	// make a clone of the out so we don't data race with other instanceSend()s
	outClone := cloneOutDest(out)

	attempts <- sendAttempt{
		outClone: outClone,
		err:      c.sendToInstance(sr, ri, fn, in, outClone),
		sp:       sp,
	}
}

func (c *ServiceClient) Send(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) (err error) {
	retry, giveup := c.GetTimeout()

	attempts := make(chan sendAttempt)

	var ticker <-chan time.Time
	if retry > 0 {
		ticker = time.NewTicker(retry).C
	}

	var timeout <-chan time.Time
	if giveup > 0 {
		timeout = time.NewTimer(giveup).C
	}

	go c.trySend(attempts, ri, fn, in, out)

	for {
		select {
		case <-ticker:
			go c.trySend(attempts, ri, fn, in, out)
		case <-timeout:
			if err == nil {
				err = ErrRequestTimeout
			}
			// otherwise use the last error reported from an attempt
			return
		case attempt := <-attempts:
			err = attempt.err
			if err == nil {
				copyOutDest(out, attempt.outClone)
				return
			}
		}
	}

	return
}

type sendAttempt struct {
	outClone interface{}
	err      error
	sp       *servicePool
}

/*
ServiceClient.SendOnce() will send a request to one of the available instances. If no response is heard after
the giveup time has passed, it will return an error.
*/
func (c *ServiceClient) SendOnce(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) (err error) {
	_, giveup := c.GetTimeout()

	attempts := make(chan sendAttempt)

	var timeout <-chan time.Time
	if giveup > 0 {
		timeout = time.NewTimer(giveup).C
	}

	go c.trySend(attempts, ri, fn, in, out)

	for {
		select {
		case <-timeout:
			if err == nil {
				err = ErrRequestTimeout
			}
			return
		case attempt := <-attempts:
			err = attempt.err
			copyOutDest(out, attempt.outClone)
			return
		}
	}

	return
}

func (c *ServiceClient) isClosed(service ServiceResource) bool {
	key := getInstanceKey(service.service)

	// TODO: this is unsafe
	if _, ok := c.instances[key]; ok {
		return false
	}

	return true
}
