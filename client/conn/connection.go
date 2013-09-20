package conn

import (
	"errors"
	"fmt"
	"github.com/kr/pretty"
	"github.com/skynetservices/skynet2"
	"github.com/skynetservices/skynet2/log"
	"github.com/skynetservices/skynet2/rpc/bsonrpc"
	"labix.org/v2/mgo/bson"
	"net"
	"net/rpc"
	"reflect"
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
	addr           string
	conn           net.Conn
	clientID       string
	serviceName    string
	rpcClient      *rpc.Client
	rpcClientCodec *bsonrpc.ClientCodec
	closed         bool

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

	cn.rpcClientCodec = bsonrpc.NewClientCodec(cn.conn)
	cn.rpcClient = rpc.NewClientWithCodec(cn.rpcClientCodec)

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

	sin := skynet.ServiceRPCInWrite{
		RequestInfo: ri,
		Method:      fn,
		ClientID:    c.clientID,
	}

	var b []byte
	b, err = bson.Marshal(in)
	if err != nil {
		return serviceError{fmt.Sprintf("Error calling bson.Marshal: %v", err)}
	}

	sin.In = bson.Binary{
		0x00,
		b,
	}

	type Resp struct {
		Out skynet.ServiceRPCOutRead
		Err error
	}

	respChan := make(chan *Resp)

	go func() {
		log.Println(log.TRACE, fmt.Sprintf("Sending Method call %s with ClientID %s to: %s", sin.Method, sin.ClientID, c.addr))
		r := &Resp{}

		r.Err = c.rpcClient.Call(c.serviceName+".Forward", sin, &r.Out)
		log.Println(log.TRACE, fmt.Sprintf("Method call %s with ClientID %s from: %s completed", sin.Method, sin.ClientID, c.addr))

		respChan <- r
	}()

	var r *Resp

	if timeout == 0 {
		timeout = 15 * time.Minute
	}

	t := time.After(timeout)

	select {
	case r = <-respChan:
		if r.Err != nil {
			err = serviceError{r.Err.Error()}
			c.Close()
			return
		}
	case <-t:
		err = fmt.Errorf("Connection: timing out request after %s", timeout.String())
		c.Close()
		return
	}

	if r.Out.ErrString != "" {
		err = serviceError{r.Out.ErrString}
		return
	}

	err = bson.Unmarshal(r.Out.Out, out)
	if err != nil {
		log.Println(log.ERROR, "Error unmarshalling nested document")
		err = serviceError{err.Error()}
		c.Close()
	}

	log.Println(log.TRACE, pretty.Sprintf("Method call %s with ClientID %s from: %s returned: %s %+v", sin.Method, sin.ClientID, c.addr, reflect.TypeOf(out), out))

	return
}

/*
Conn.performHandshake Responsible for performing handshake with service
*/
func (c *Conn) performHandshake() (err error) {
	var sh skynet.ServiceHandshake
	log.Println(log.TRACE, "Reading ServiceHandshake")

	err = c.rpcClientCodec.Decoder.Decode(&sh)
	if err != nil {
		log.Println(log.ERROR, "Failed to decode ServiceHandshake", err)
		c.conn.Close()

		return HandshakeFailed
	}

	if sh.Name != c.serviceName {
		log.Println(log.ERROR, "Attempted to send request to incorrect service: "+sh.Name)
		c.conn.Close()
		return HandshakeFailed
	}

	ch := skynet.ClientHandshake{}

	log.Println(log.TRACE, "Writing ClientHandshake")
	err = c.rpcClientCodec.Encoder.Encode(ch)
	if err != nil {
		log.Println(log.ERROR, "Failed to encode ClientHandshake", err)
		c.conn.Close()

		return HandshakeFailed
	}

	if !sh.Registered {
		log.Println(log.ERROR, "Attempted to send request to unregistered service")
		return ServiceUnregistered
	}

	log.Println(log.TRACE, "Handing connection RPC layer")
	c.clientID = sh.ClientID

	return
}
