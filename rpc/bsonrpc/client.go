package bsonrpc

import (
	"errors"
	"github.com/kr/pretty"
	"github.com/skynetservices/skynet/log"
	"io"
	"net/rpc"
	"reflect"
)

type ClientCodec struct {
	conn    io.ReadWriteCloser
	Encoder *Encoder
	Decoder *Decoder
}

func NewClientCodec(conn io.ReadWriteCloser) (codec *ClientCodec) {
	cc := &ClientCodec{
		conn:    conn,
		Encoder: NewEncoder(conn),
		Decoder: NewDecoder(conn),
	}
	codec = cc
	return
}

func (cc *ClientCodec) WriteRequest(req *rpc.Request, v interface{}) (err error) {
	log.Println(log.TRACE, "RPC Client Entered: WriteRequest")
	defer log.Println(log.TRACE, "RPC Client Leaving: WriteRequest")

	log.Println(log.TRACE, pretty.Sprintf("RPC Client Writing RequestHeader %s %+v", reflect.TypeOf(req), req))

	err = cc.Encoder.Encode(req)
	if err != nil {
		log.Println(log.ERROR, "RPC Client Error enconding request rpc request: ", err)
		cc.Close()
		return
	}

	log.Println(log.TRACE, pretty.Sprintf("RPC Client Writing Request Value %s %+v", reflect.TypeOf(v), v))

	err = cc.Encoder.Encode(v)
	if err != nil {
		log.Println(log.ERROR, "RPC Client Error enconding request value: ", err)
		cc.Close()
		return
	}

	return
}

func (cc *ClientCodec) ReadResponseHeader(res *rpc.Response) (err error) {
	log.Println(log.TRACE, "RPC Client Entered: ReadResponseHeader")
	defer log.Println(log.TRACE, "RPC Client Leaving: ReadResponseHeader")

	err = cc.Decoder.Decode(res)

	if err != nil {
		cc.Close()
		log.Println(log.ERROR, "RPC Client Error decoding response header: ", err)
	}

	if err == nil {
		log.Println(log.TRACE, pretty.Sprintf("RPC Client Read ResponseHeader %s %+v", reflect.TypeOf(res), res))
	}

	return
}

func (cc *ClientCodec) ReadResponseBody(v interface{}) (err error) {
	log.Println(log.TRACE, "RPC Client Entered: ReadResponseBody")
	defer log.Println(log.TRACE, "RPC Client Leaving: ReadResponseBody")

	if v == nil {
		err = errors.New("Response object cannot be nil")
		log.Println(log.ERROR, "RPC Client Error reading response body: ", err)
		return
	}

	err = cc.Decoder.Decode(v)

	if err != nil {
		cc.Close()
		log.Println(log.ERROR, "RPC Client Error decoding response body: ", err)
	}

	if err == nil {
		log.Println(log.TRACE, pretty.Sprintf("RPC Client Read ResponseBody %s %+v", reflect.TypeOf(v), v))
	}
	return
}

func (cc *ClientCodec) Close() (err error) {
	log.Println(log.TRACE, "RPC Client Entered: Close")
	defer log.Println(log.TRACE, "RPC Client Leaving: Close")

	err = cc.conn.Close()

	if err != nil && err.Error() != "use of closed network connection" {
		log.Println(log.ERROR, "RPC Client Error closing connection: ", err)
	}

	return
}

func NewClient(conn io.ReadWriteCloser) (c *rpc.Client) {
	cc := NewClientCodec(conn)
	c = rpc.NewClientWithCodec(cc)
	return
}
