package bsonrpc

import (
	"io"
	"net/rpc"
)

type scodec struct {
	conn io.ReadWriteCloser
	enc  *Encoder
	dec  *Decoder
}

func NewServerCodec(conn io.ReadWriteCloser) (codec rpc.ServerCodec) {
	sc := &scodec{
		conn: conn,
		enc:  NewEncoder(conn),
		dec:  NewDecoder(conn),
	}
	codec = sc
	return
}

func (sc *scodec) ReadRequestHeader(rq *rpc.Request) (err error) {
	err = sc.dec.Decode(rq)
	return
}

func (sc *scodec) ReadRequestBody(v interface{}) (err error) {
	err = sc.dec.Decode(v)
	return
}

func (sc *scodec) WriteResponse(rs *rpc.Response, v interface{}) (err error) {
	err = sc.enc.Encode(rs)
	if err != nil {
		return
	}
	err = sc.enc.Encode(v)
	if err != nil {
		return
	}
	return
}

func (sc *scodec) Close() (err error) {
	err = sc.conn.Close()
	return
}

func ServeConn(conn io.ReadWriteCloser) (s *rpc.Server) {
	s = rpc.NewServer()
	s.ServeCodec(NewServerCodec(conn))
	return
}
