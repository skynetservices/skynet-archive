package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/4ad/doozer"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/pools"
	"github.com/bketelsen/skynet/rpc/bsonrpc"
	"github.com/bketelsen/skynet/service"
	"launchpad.net/mgo/v2/bson"
	"math/rand"
	"net"
	"path"
	"strings"
)

type ServiceClient struct {
	Log     skynet.Logger `json:"-"`
	cconfig *skynet.ClientConfig
	//connectionPool *pools.RoundRobin
	query     *Query
	instances map[string]servicePool
	muxChan   chan interface{}
}

func newServiceClient(query *Query, c *Client) (sc *ServiceClient) {
	sc = &ServiceClient{
		Log:       c.Config.Log,
		cconfig:   c.Config,
		query:     query,
		instances: make(map[string]servicePool, 0),
		muxChan:   make(chan interface{}),
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

func getConnectionFactory(s *service.Service) (factory pools.Factory) {
	factory = func() (pools.Resource, error) {
		conn, err := net.Dial("tcp", s.Config.ServiceAddr.String())

		if err != nil {
			// TODO: handle failure here and attempt to connect to a different instance
			return nil, errors.New("Failed to connect to service: " + s.Config.ServiceAddr.String())
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
			return factory()
		}

		resource := ServiceResource{
			rpcClient: bsonrpc.NewClient(conn),
			service:   *s,
		}

		return resource, nil
	}
	return
}

type servicePool struct {
	service *service.Service
	pool    *pools.ResourcePool
}

type lightInstanceRequest struct {
	response chan servicePool
}

func (c *ServiceClient) mux() {
	var spSubscribers []chan servicePool

	for {
		select {
		case mi := <-c.muxChan:
			switch m := mi.(type) {
			case service.ServiceDiscovered:
				sp := servicePool{
					service: m.Service,
					pool:    pools.NewResourcePool(getConnectionFactory(m.Service), c.cconfig.ConnectionPoolSize, c.cconfig.ConnectionPoolSize),
				}
				_, known := c.instances[m.Service.Config.ServiceAddr.String()]
				c.instances[m.Service.Config.ServiceAddr.String()] = sp
				if !known {
					c.Log.Item(m)
				}
				// send this instance to anyone who was waiting
				for _, sps := range spSubscribers {
					sps <- sp
				}
				// no one is waiting anymore
				spSubscribers = spSubscribers[:0]
			case service.ServiceRemoved:
				delete(c.instances, m.Service.Config.ServiceAddr.String())
				c.Log.Item(m)
			case lightInstanceRequest:
				sp, ok := c.getLightInstanceMux()
				if ok {
					m.response <- sp
				} else {
					//if one wasn't immediately available, wait for the next incoming
					spSubscribers = append(spSubscribers, m.response)
				}
			}
		}
	}
}

// do not call this from outside .mux()
func (c *ServiceClient) getLightInstanceMux() (sp servicePool, ok bool) {
	if len(c.instances) == 0 {
		ok = false
		return
	}

	// first collect those that have the greatest reported number of available slots
	// mostSlots := 0
	bestInstances := make([]servicePool, len(c.instances), 0)
	for _, i := range c.instances {
		// let's just add them all for the moment
		/*
			if i.service.Slots > mostSlots {
				mostSlots = i.service.Slots
				bestInstances = bestInstances[:0]
			}
			if i.service.Slots < mostSlots {
				continue
			}
		*/
		bestInstances = append(bestInstances, i)
	}

	// then choose one randomly

	ri := rand.Intn(len(bestInstances))
	sp = bestInstances[ri]
	ok = true

	return
}

func (c *ServiceClient) getLightInstance() (sp servicePool) {
	response := make(chan servicePool, 1)
	c.muxChan <- lightInstanceRequest{response}
	sp = <-response
	return
}

func (c *ServiceClient) Send(requestInfo *skynet.RequestInfo, funcName string, in interface{}, outPointer interface{}) (err error) {
	// TODO: timeout logic
	s, sp, err := c.getConnection(0)
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

	sp.pool.Release(s)

	return
}

func (c *ServiceClient) getConnection(lvl int) (service ServiceResource, sp servicePool, err error) {
	if lvl > 5 {
		err = errors.New("Unable to retrieve a valid connection to the service")
		return
	}

	sp = c.getLightInstance()

	conn, err := sp.pool.Acquire()

	if err != nil || c.isClosed(conn.(ServiceResource)) {
		if conn != nil {
			s := conn.(ServiceResource)

			s.closed = true
			sp.pool.Release(s)
		}

		return c.getConnection(lvl + 1)
	}

	service = conn.(ServiceResource)

	return
}

func (c *ServiceClient) isClosed(service ServiceResource) bool {
	key := getInstanceKey(service.service)

	if _, ok := c.instances[key]; ok {
		return false
	}

	return true
}
