package client

import (
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/rpc/bsonrpc"
	"net"
)

type Admin struct {
	Instance *skynet.ServiceInfo
}

func (a *Admin) Register(in skynet.RegisterRequest) (out skynet.RegisterResponse, err error) {
	conn, err := net.Dial("tcp", a.Instance.Config.AdminAddr.String())
	if err != nil {
		return
	}
	rpcClient := bsonrpc.NewClient(conn)
	err = rpcClient.Call("Admin.Register", in, &out)
	rpcClient.Close()
	return
}

func (a *Admin) Unregister(in skynet.UnregisterRequest) (out skynet.UnregisterResponse, err error) {
	conn, err := net.Dial("tcp", a.Instance.Config.AdminAddr.String())
	if err != nil {
		return
	}
	rpcClient := bsonrpc.NewClient(conn)
	err = rpcClient.Call("Admin.UnregisterResponse", in, &out)
	rpcClient.Close()
	return
}

func (a *Admin) Stop(in skynet.StopRequest) (out skynet.StopResponse, err error) {
	conn, err := net.Dial("tcp", a.Instance.Config.AdminAddr.String())
	if err != nil {
		return
	}
	rpcClient := bsonrpc.NewClient(conn)
	err = rpcClient.Call("Admin.Stop", in, &out)
	rpcClient.Close()
	return
}
