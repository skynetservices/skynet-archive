package client

import (
	"github.com/bketelsen/skynet/rpc/bsonrpc"
	"github.com/bketelsen/skynet/service"
	"net"
)

type Admin struct {
	Instance *service.Service
}

func (a *Admin) Register(in service.RegisterRequest) (out service.RegisterResponse, err error) {
	conn, err := net.Dial("tcp", a.Instance.Config.AdminAddr.String())
	if err != nil {
		return
	}
	rpcClient := bsonrpc.NewClient(conn)
	err = rpcClient.Call("Admin.Register", in, &out)
	rpcClient.Close()
	return
}

func (a *Admin) Unregister(in service.UnregisterRequest) (out service.UnregisterResponse, err error) {
	conn, err := net.Dial("tcp", a.Instance.Config.AdminAddr.String())
	if err != nil {
		return
	}
	rpcClient := bsonrpc.NewClient(conn)
	err = rpcClient.Call("Admin.UnregisterResponse", in, &out)
	rpcClient.Close()
	return
}

func (a *Admin) Stop(in service.StopRequest) (out service.StopResponse, err error) {
	conn, err := net.Dial("tcp", a.Instance.Config.AdminAddr.String())
	if err != nil {
		return
	}
	rpcClient := bsonrpc.NewClient(conn)
	err = rpcClient.Call("Admin.Stop", in, &out)
	rpcClient.Close()
	return
}
