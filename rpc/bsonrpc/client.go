package bsonrpc

import (
	"bufio"
	"errors"
	"io"
	"net/rpc"
)

type ccodec struct {
	conn   io.ReadWriteCloser
	enc    *Encoder
	dec    *Decoder
	encBuf *bufio.Writer
}

func NewClientCodec(conn io.ReadWriteCloser) (codec rpc.ClientCodec) {
	encBuf := bufio.NewWriter(conn)
	cc := &ccodec{
		conn:   conn,
		enc:    NewEncoder(encBuf),
		dec:    NewDecoder(conn),
		encBuf: encBuf,
	}
	codec = cc
	return
}

func (cc *ccodec) WriteRequest(req *rpc.Request, v interface{}) (err error) {
	err = cc.enc.Encode(req)
	if err != nil {
		return
	}
	err = cc.enc.Encode(v)
	if err != nil {
		return
	}
	return cc.encBuf.Flush()
}

func (cc *ccodec) ReadResponseHeader(res *rpc.Response) (err error) {
	err = cc.dec.Decode(res)
	return
}

func (cc *ccodec) ReadResponseBody(v interface{}) (err error) {
	if v == nil {
		return errors.New("Response object cannot be nil")
	}

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
