package conn

import (
	"errors"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/rpc/bsonrpc"
	"labix.org/v2/mgo/bson"
	"net"
	"net/rpc"
	"testing"
	"time"
)

// TODO: One of these tests is bailing early and occassionaly causing EOF issues with BSON
func TestHandshake(t *testing.T) {
	client, server := net.Pipe()

	go doServiceHandshake(server, true, t)

	cn, err := NewConnectionFromNetConn("TestService", client)
	c := cn.(*Conn)

	if err != nil {
		t.Fatal("Failed to perform handshake", err)
	}

	if c.rpcClient == nil {
		t.Fatal("rpc.Client not initialized")
	}

	c.Close()
	server.Close()
}

func TestErrorOnUnregistered(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	go doServiceHandshake(server, false, t)

	_, err := NewConnectionFromNetConn("TestService", client)

	if err != ServiceUnregistered {
		t.Fatal("Connection should return error when service is unregistered")
	}
}

func TestDialConn(t *testing.T) {
	ln, err := net.Listen("tcp", ":51900")
	defer ln.Close()

	if err != nil {
		t.Error("Failed to bind to port for test")
	}

	go func() {
		conn, err := ln.Accept()
		if err == nil {
			go doServiceHandshake(conn, true, t)
		}
	}()

	c, err := NewConnection("TestService", "tcp", ":51900", 500*time.Millisecond)

	if err != nil || c == nil {
		t.Fatal("NewConnection() failed to establish tcp connection", err)
	}
}

func TestSetIdleTimeout(t *testing.T) {
	client, _ := net.Pipe()
	defer client.Close()

	c := Conn{conn: client}
	c.SetIdleTimeout(1 * time.Minute)

	if c.idleTimeout != 1*time.Minute {
		t.Fatal("IdleTimeout not set as expected")
	}
}

func TestSendTimeout(t *testing.T) {
	ln, err := net.Listen("tcp", ":51900")
	defer ln.Close()

	if err != nil {
		t.Error("Failed to bind to port for test")
	}

	go func() {
		conn, err := ln.Accept()
		if err == nil {
			doServiceHandshake(conn, true, t)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	c, err := NewConnection("TestService", "tcp", ":51900", 500*time.Millisecond)

	if err != nil {
		t.Error("Failed to establish connection for test", err)
	}

	var o string
	err = c.SendTimeout(&skynet.RequestInfo{}, "foo", 10, &o, 2*time.Millisecond)

	if err == nil {
		t.Fatal("Expected SendTimeout to return timeout error")
	}
}

func TestSend(t *testing.T) {
	client, server := net.Pipe()
	go doServiceHandshake(server, true, t)

	cn, err := NewConnectionFromNetConn("TestRPCService", client)
	c := cn.(*Conn)

	s := rpc.NewServer()
	var ts TestRPCService
	s.Register(&ts)
	go s.ServeCodec(bsonrpc.NewServerCodec(server))

	var tp TestParam
	tp.Val1 = "Hello World"
	tp.Val2 = 10

	ri := &skynet.RequestInfo{}

	ts.TestMethod = func(in skynet.ServiceRPCIn, out *skynet.ServiceRPCOut) (err error) {
		out.Out, err = bson.Marshal(&tp)

		var t TestParam

		if err != nil {
			return
		}

		if in.ClientID != c.clientID {
			return errors.New("Failed to set ClientID on request")
		}

		if in.Method != "Foo" {
			return errors.New("Failed to set Method on request")
		}

		if *in.RequestInfo != *ri {
			return errors.New("Failed to set RequestInfo on request")
		}

		err = bson.Unmarshal(in.In, &t)
		if err != nil {
			return
		}

		if t.Val1 != tp.Val1 || tp.Val2 != tp.Val2 {
			return errors.New("Request failed to send proper data")
		}

		return
	}

	err = c.Send(ri, "Foo", tp, &tp)
	if err != nil {
		t.Error(err)
		return
	}

	c.Close()
	server.Close()
}

func TestSendOnClosedConnection(t *testing.T) {
	client, server := net.Pipe()
	go doServiceHandshake(server, true, t)

	c, err := NewConnectionFromNetConn("TestService", client)
	c.Close()

	var tp TestParam
	tp.Val1 = "Hello World"
	tp.Val2 = 10

	ri := &skynet.RequestInfo{}

	err = c.Send(ri, "foo", tp.Val1, &tp.Val2)

	if err != ConnectionClosed {
		t.Fatal("Send() should not send when connection has been closed")
	}
}

/*
* Test Helpers
 */

type TestParam struct {
	Val1 string
	Val2 int
}

type TestRPCService struct {
	TestMethod func(in skynet.ServiceRPCIn, out *skynet.ServiceRPCOut) (err error)
}

func (ts *TestRPCService) Forward(in skynet.ServiceRPCIn, out *skynet.ServiceRPCOut) (err error) {
	if ts.TestMethod != nil {
		return ts.TestMethod(in, out)
	}

	return errors.New("No Method Supplied")
}

func (ts TestRPCService) Foo(in TestParam, out *TestParam) (err error) {
	out.Val1 = in.Val1 + "world!"
	out.Val2 = in.Val2 + 5
	return
}

func doServiceHandshake(server net.Conn, registered bool, t *testing.T) {
	sh := skynet.ServiceHandshake{
		Registered: registered,
		ClientID:   "abc",
	}

	encoder := bsonrpc.NewEncoder(server)
	err := encoder.Encode(sh)
	if err != nil {
		t.Fatal("Failed to encode server handshake", err)
	}

	var ch skynet.ClientHandshake
	decoder := bsonrpc.NewDecoder(server)
	err = decoder.Decode(&ch)
	if err != nil {
		t.Fatal("Error calling bsonrpc.NewDecoder: ", err)
	}
}
