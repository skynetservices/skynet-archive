package bsonrpc

import (
	"io"
	"net/rpc"
)

/*
type ClientCodec interface {
    WriteRequest(*Request, interface{}) error
    ReadResponseHeader(*Response) error
    ReadResponseBody(interface{}) error

    Close() error
}
*/

type ccodec struct {
	conn io.ReadWriteCloser
	enc  *Encoder
	dec  *Decoder
}

func NewClientCodec(conn io.ReadWriteCloser) (codec rpc.ClientCodec) {
	cc := &ccodec{
		conn: conn,
		enc:  NewEncoder(conn),
		dec:  NewDecoder(conn),
	}
	codec = cc
	return
}

/*
type Request struct {
	ServiceMethod string   // format: "Service.Method"
	Seq           uint64   // sequence number chosen by client
	next          *Request // for free list in Server
}

type Response struct {
	ServiceMethod string    // echoes that of the Request
	Seq           uint64    // echoes that of the request
	Error         string    // error, if any.
	next          *Response // for free list in Server
}
*/

func (cc *ccodec) WriteRequest(req *rpc.Request, v interface{}) (err error) {
	err = cc.enc.Encode(req)
	if err != nil {
		return
	}
	err = cc.enc.Encode(v)
	if err != nil {
		return
	}
	return
}

func (cc *ccodec) ReadResponseHeader(res *rpc.Response) (err error) {
	err = cc.dec.Decode(res)
	return
}

func (cc *ccodec) ReadResponseBody(v interface{}) (err error) {
	err = cc.dec.Decode(v)
	return
}

func (cc *ccodec) Close() (err error) {
	err = cc.conn.Close()
	return
}

func NewClient(conn io.ReadWriteCloser) (c *rpc.Client) {
	cc := NewClientCodec(conn)
	c = rpc.NewClientWithCodec(cc)
	return
}
