package skylib

import (
	"fmt"
	"testing"
)

type M map[string]interface{}

type EchoRPC struct {
}

func (e EchoRPC) Started(s *Service)      {}
func (e EchoRPC) Stopped(s *Service)      {}
func (e EchoRPC) Registered(s *Service)   {}
func (e EchoRPC) Unregistered(s *Service) {}

func (e EchoRPC) Foo(rinfo RequestInfo, in M, out *M) (err error) {
	*out = M{
		"Hi": in["Hi"],
	}
	return
}

func TestServiceRPCBasic(t *testing.T) {
	srpc := NewServiceRPC(EchoRPC{})
	in := M{"Hi": "there"}
	out := &M{}

	sin := ServiceRPCIn{
		RequestInfo: RequestInfo{RequestID: "id"},
		Method:      "Foo",
		In:          in,
	}
	sout := ServiceRPCOut{
		Out: out,
	}

	err := srpc.Forward(sin, &sout)
	if err != nil {
		t.Error(err)
	}
	if v, ok := (*out)["Hi"].(string); !ok || v != "there" {
		t.Error(fmt.Sprintf("Expected %v, got %v", in, *out))
	}
}
