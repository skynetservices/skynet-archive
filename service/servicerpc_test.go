package service

import (
	"fmt"
	"github.com/skynetservices/mgo/bson"
	"github.com/skynetservices/skynet"
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

//define StatsdStub type
type StatsdStub struct {
	TimingFunc func(stat string, value int64, rate float32) error
	CloseFunc  func() error
}

//implement the functions to comply with StatsdClient interface
func (sd *StatsdStub) Timing(stat string, value int64, rate float32) error {
	if sd.TimingFunc != nil {
		return sd.TimingFunc(stat, value, rate)
	}
	return nil
}

func (sd *StatsdStub) Close() error {
	return nil
}

func TestServiceRPCBasic(t *testing.T) {
	fmt.Println("TestServiceRPCBasic")

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

func TestStatsd(t *testing.T) {
	fmt.Println("TestStatsd")
	var addr net.Addr

	timingCalled := false

	stub := &StatsdStub{
		TimingFunc: func(stat string, duration int64, rate float32) (err error) {

			timingCalled = true

			if stat != "duration" {
				t.Fatal("Wrong stat, expected duration, got " + stat)
			}
			//I don't think I can make assertion for duration??
			if rate != 1.0 {
				t.Fatal("Wrong rate, expected 1.0, got %+v", rate)
			}
			return
		},
	}

	config := &skynet.ServiceConfig{}
	//StatsCfg fields are set through the flags, if they are not set
	//statsdClient is not initialized and not called.
	config.StatsCfg = new(skynet.StatsdConfig)
	config.StatsCfg.Addr = "127.0.0.1:8125"
	config.StatsCfg.Dir = "default"

	service := CreateService(EchoRPC{}, config)
	service.clientInfo = make(map[string]ClientInfo, 1)

	service.statsdClient = stub

	addr = &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 123,
	}

	service.clientInfo["123"] = ClientInfo{
		Address: addr,
	}

	srpc := NewServiceRPC(service)

	in := M{"Hi": "there"}
	//	out := &M{}

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

	if !timingCalled {
		t.Error("StatsD Timing function was not called by ServiceRPC::Forward")
	}

}
