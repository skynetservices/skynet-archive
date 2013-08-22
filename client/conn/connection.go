package conn

import (
	"errors"
	"fmt"
	"github.com/skynetservices/skynet2"
	"github.com/skynetservices/skynet2/log"
	"github.com/skynetservices/skynet2/rpc/bsonrpc"
	"labix.org/v2/mgo/bson"
	"net"
	"net/rpc"
	"time"
)

// TODO: Abstract out BSON logic into an interface that can be proviced for Encoding/Decoding data to a supplied interface
// this would allow developers to swap out the RPC logic, maybe implement our own ClientCodec/ServerCodec that have an additional WriteHandshake/ReadHandshake methods on each of them.
// bson for example we could create a custom type that is composed of our methods and the normal rpc codec
var (
	HandshakeFailed     = errors.New("Handshake Failed")
	ServiceUnregistered = errors.New("Service is unregistered")
	ConnectionClosed    = errors.New("Connection is closed")
)

type serviceError struct {
	msg string
}

func (se serviceError) Error() string {
	return se.msg
}

/*
Connection
*/

type Connection interface {
	SetIdleTimeout(timeout time.Duration)
	Addr() string

	Close()
	IsClosed() bool

	Send(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) (err error)
	SendTimeout(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}, timeout time.Duration) (err error)
}

/*
Conn
Implementation of Connection
*/
type Conn struct {
	addr        string
	conn        net.Conn
	clientID    string
	serviceName string
	rpcClient   *rpc.Client
	closed      bool

	idleTimeout time.Duration
}

/*
client.NewConnection() Establishes new connection to skynet service specified by addr
*/
func NewConnection(serviceName, network, addr string, timeout time.Duration) (conn Connection, err error) {
	c, err := net.DialTimeout(network, addr, timeout)

	if err != nil {
		return
	}

	conn, err = NewConnectionFromNetConn(serviceName, c)

	return
}

/*
client.NewConn() Establishes new connection to skynet service with existing net.Conn
This is beneficial if you want to communicate over a pipe
*/
func NewConnectionFromNetConn(serviceName string, c net.Conn) (conn Connection, err error) {
	cn := &Conn{conn: c}
	cn.addr = c.RemoteAddr().String()
	cn.serviceName = serviceName

	err = cn.performHandshake()

	return cn, err
}

/*
Conn.Close() Close network connection
*/
func (c *Conn) Close() {
	c.closed = true
	c.rpcClient.Close()
}

/*
Conn.SetIdleTimeout() amount of time that can pass between requests before connection is closed
*/
func (c *Conn) SetIdleTimeout(timeout time.Duration) {
	c.idleTimeout = timeout
	c.setDeadline(timeout)
}

/*
Conn.IsClosed() Specifies if connection is closed
*/
func (c Conn) IsClosed() bool {
	return c.closed
}

/*
Conn.Addr() Specifies the network address
*/
func (c Conn) Addr() string {
	return c.addr
}

/*
Conn.Send() Sends RPC request to service
*/
func (c *Conn) Send(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}) (err error) {
	return c.SendTimeout(ri, fn, in, out, 0)
}

/*
Conn.SendTimeout() Acts like Send but takes a timeout
*/
func (c *Conn) SendTimeout(ri *skynet.RequestInfo, fn string, in interface{}, out interface{}, timeout time.Duration) (err error) {
	if c.IsClosed() {
		return ConnectionClosed
	}

	sin := skynet.ServiceRPCIn{
		RequestInfo: ri,
		Method:      fn,
		ClientID:    c.clientID,
	}

	sin.In, err = bson.Marshal(in)
	if err != nil {
		return serviceError{fmt.Sprintf("Error calling bson.Marshal: %v", err)}
	}

	sout := skynet.ServiceRPCOut{}

	// Set timeout for this request, then set it back to idle timeout
	c.setDeadline(timeout)
	defer c.setDeadline(c.idleTimeout)

	err = c.rpcClient.Call(c.serviceName+".Forward", sin, &sout)
	if err != nil {
		c.Close()
		err = serviceError{err.Error()}

		return
	}

	if sout.ErrString != "" {
		err = serviceError{sout.ErrString}
		return
	}

	err = bson.Unmarshal(sout.Out, out)
	if err != nil {
		err = serviceError{err.Error()}
	}

	return
}

func (c *Conn) setDeadline(timeout time.Duration) {
	if timeout == 0 {
		var t time.Time
		c.conn.SetDeadline(t)
	} else {
		c.conn.SetDeadline(time.Now().Add(timeout))
	}
}

/*
Conn.performHandshake Responsible for performing handshake with service
*/
func (c *Conn) performHandshake() (err error) {
	var sh skynet.ServiceHandshake
	decoder := bsonrpc.NewDecoder(c.conn)

	err = decoder.Decode(&sh)
	if err != nil {
		log.Println(log.ERROR, "Failed to decode ServiceHandshake", err)
		c.conn.Close()

		return HandshakeFailed
	}

	ch := skynet.ClientHandshake{}
	encoder := bsonrpc.NewEncoder(c.conn)

	err = encoder.Encode(ch)
	if err != nil {
		log.Println(log.ERROR, "Failed to encode ClientHandshake", err)
		c.conn.Close()

		return HandshakeFailed
	}

	if !sh.Registered {
		return ServiceUnregistered
	}

	c.rpcClient = bsonrpc.NewClient(c.conn)
	c.clientID = sh.ClientID

	return
}
