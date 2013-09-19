package bsonrpc

import (
	"bufio"
	"github.com/skynetservices/skynet2/log"
	"io"
	"net/rpc"
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
	err = sc.dec.Decode(rq)
	if err != nil && err != io.EOF {
		log.Println(log.ERROR, "RPC Server Error decoding request header: ", err)
	}
	return
}

func (sc *scodec) ReadRequestBody(v interface{}) (err error) {
	err = sc.dec.Decode(v)
	if err != nil {
		log.Println(log.ERROR, "RPC Server Error decoding request body: ", err)
	}
	return
}

func (sc *scodec) WriteResponse(rs *rpc.Response, v interface{}) (err error) {
	err = sc.enc.Encode(rs)
	if err != nil {
		log.Println(log.ERROR, "RPC Server Error encoding rpc response: ", err)
		return
	}
	err = sc.enc.Encode(v)
	if err != nil {
		log.Println(log.ERROR, "RPC Server Error encoding response value: ", err)
		return
	}
	return sc.encBuf.Flush()
}

func (sc *scodec) Close() (err error) {
	err = sc.conn.Close()
	if err != nil {
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
