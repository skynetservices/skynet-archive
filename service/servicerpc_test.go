package service

import (
	"fmt"
	"github.com/bketelsen/skynet"
	"labix.org/v2/mgo/bson"
	"net"
	"testing"
)

type M map[string]interface{}

type EchoRPC struct {
}

func (e EchoRPC) Started(s *Service)      {}
func (e EchoRPC) Stopped(s *Service)      {}
func (e EchoRPC) Registered(s *Service)   {}
func (e EchoRPC) Unregistered(s *Service) {}

func (e EchoRPC) Foo(rinfo *skynet.RequestInfo, in M, out *M) (err error) {
	*out = M{
		"Hi": in["Hi"],
	}

	return
}

func TestServiceRPCBasic(t *testing.T) {
	var addr net.Addr

	config := &skynet.ServiceConfig{}
	service := CreateService(EchoRPC{}, config)
	service.clientInfo = make(map[string]ClientInfo, 1)

	addr = &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 123,
	}

	service.clientInfo["123"] = ClientInfo{
		Address: addr,
	}

	srpc := NewServiceRPC(service)

	in := M{"Hi": "there"}
	out := &M{}

	sin := skynet.ServiceRPCIn{
		RequestInfo: &skynet.RequestInfo{
			RequestID:         "id",
			OriginAddress:     addr.String(),
			ConnectionAddress: addr.String(),
		},
		Method:   "Foo",
		ClientID: "123",
	}

	sin.In, _ = bson.Marshal(in)

	sout := skynet.ServiceRPCOut{}

	err := srpc.Forward(sin, &sout)
	if err != nil {
		t.Error(err)
	}

	bson.Unmarshal(sout.Out, out)

	if v, ok := (*out)["Hi"].(string); !ok || v != "there" {
		t.Error(fmt.Sprintf("Expected %v, got %v", in, *out))
	}
}
