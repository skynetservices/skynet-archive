package bsonrpc

import (
	"io"
	"testing"
	"net/rpc"
)

type duplex struct {
	io.Reader
	io.Writer
}

func (d duplex) Close() (err error) {
	return
}

type TestParam struct {
	Val1 string
	Val2 int
}

type Test int

func (ts Test) Foo(in TestParam, out *TestParam) (err error) {
	out.Val1 = in.Val1 + "world!"
	out.Val2 = in.Val2 + 5
	return
}

func basicServer(conn io.ReadWriteCloser) {
	s := ServeConn(conn)
	var ts Test
	s.Register(&ts)
}

func TestBasicClientServer(t *testing.T) {
	toServer, fromClient := io.Pipe()
	toClient, fromServer := io.Pipe()

	s := rpc.NewServer()
	var ts Test
	s.Register(&ts)
	go s.ServeCodec(NewServerCodec(duplex{toServer, fromServer}))

	cl := NewClient(duplex{toClient, fromClient})

	var tp TestParam
	tp.Val1 = "Hello "
	tp.Val2 = 10

	err := cl.Call("Test.Foo", tp, &tp)
	if err != nil {
		t.Error(err)
		return
	}
	if tp.Val1 != "Hello world!" {
		t.Errorf("tp.Val2: expected %q, got %q", "Hello world!", tp.Val1)
	}
	if tp.Val2 != 15 {
		t.Errorf("tp.Val2: expected 15, got %d", tp.Val2)
	}
}
