package bsonrpc

import (
	"bufio"
	"github.com/kr/pretty"
	"github.com/skynetservices/skynet2/log"
	"io"
	"net/rpc"
	"reflect"
)

type scodec struct {
	conn   io.ReadWriteCloser
	enc    *Encoder
	dec    *Decoder
	encBuf *bufio.Writer
}

func NewServerCodec(conn io.ReadWriteCloser) (codec rpc.ServerCodec) {
	encBuf := bufio.NewWriter(conn)
	sc := &scodec{
		conn:   conn,
		enc:    NewEncoder(encBuf),
		dec:    NewDecoder(conn),
		encBuf: encBuf,
	}
	codec = sc
	return
}

func (sc *scodec) ReadRequestHeader(rq *rpc.Request) (err error) {
	log.Println(log.TRACE, "RPC Server Entered: ReadRequestHeader")
	defer log.Println(log.TRACE, "RPC Server Leaving: ReadRequestHeader")

	err = sc.dec.Decode(rq)
	if err != nil && err != io.EOF {
		log.Println(log.ERROR, "RPC Server Error decoding request header: ", err)
		sc.Close()
	}

	if err == nil {
		log.Println(log.TRACE, pretty.Sprintf("RPC Server Read RequestHeader %s %+v", reflect.TypeOf(rq), rq))
	}
	return
}

func (sc *scodec) ReadRequestBody(v interface{}) (err error) {
	log.Println(log.TRACE, "RPC Server Entered: ReadRequestBody")
	defer log.Println(log.TRACE, "RPC Server Leaving: ReadRequestBody")

	err = sc.dec.Decode(v)
	if err != nil {
		log.Println(log.ERROR, "RPC Server Error decoding request body: ", err)
	}

	if err == nil {
		log.Println(log.TRACE, pretty.Sprintf("RPC Server Read RequestBody %s %+v", reflect.TypeOf(v), v))
	}
	return
}

func (sc *scodec) WriteResponse(rs *rpc.Response, v interface{}) (err error) {
	log.Println(log.TRACE, "RPC Server Entered: WriteResponse")
	defer log.Println(log.TRACE, "RPC Server Leaving: WriteResponse")

	log.Println(log.TRACE, pretty.Sprintf("RPC Server Writing ResponseHeader %s %+v", reflect.TypeOf(rs), rs))

	err = sc.enc.Encode(rs)
	if err != nil {
		log.Println(log.ERROR, "RPC Server Error encoding rpc response: ", err)
		sc.Close()
		return
	}

	log.Println(log.TRACE, pretty.Sprintf("RPC Server Writing Response Value %s %+v", reflect.TypeOf(v), v))

	err = sc.enc.Encode(v)
	if err != nil {
		log.Println(log.ERROR, "RPC Server Error encoding response value: ", err)
		sc.Close()
		return
	}
	return sc.encBuf.Flush()
}

func (sc *scodec) Close() (err error) {
	log.Println(log.TRACE, "RPC Server Entered: Close")
	defer log.Println(log.TRACE, "RPC Server Leaving: Close")

	err = sc.conn.Close()
	if err != nil && err.Error() != "use of closed network connection" {
		log.Println(log.ERROR, "RPC Server Error closing connection: ", err)
		return
	}
	return
}

func ServeConn(conn io.ReadWriteCloser) (s *rpc.Server) {
	s = rpc.NewServer()
	s.ServeCodec(NewServerCodec(conn))
	return
}
