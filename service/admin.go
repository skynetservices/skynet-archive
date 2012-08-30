package service

import (
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/rpc/bsonrpc"
	"net/rpc"
)

type ServiceAdmin struct {
	service *Service
	rpc     *rpc.Server
}

func NewServiceAdmin(service *Service) (sa *ServiceAdmin) {
	sa = &ServiceAdmin{
		service: service,
		rpc:     rpc.NewServer(),
	}

	sa.rpc.Register(&Admin{
		service: service,
	})

	return
}

func (sa *ServiceAdmin) Listen(addr *skynet.BindAddr) {
	listener, err := addr.Listen()
	if err != nil {
		panic(err)
	}

	sa.service.Log.Item(AdminListening{sa.service.Config})

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			panic(err)
		}
		go sa.rpc.ServeCodec(bsonrpc.NewServerCodec(conn))
	}
}

type Admin struct {
	service *Service
}

type RegisterRequest struct {
}

type RegisterResponse struct {
}

func (sa *Admin) Register(in RegisterRequest, out *RegisterResponse) (err error) {
	sa.service.Log.Println("Got RPC admin command Register")
	sa.service.Register()
	return
}

type UnregisterRequest struct {
}

type UnregisterResponse struct {
}

func (sa *Admin) Unregister(in UnregisterRequest, out *UnregisterResponse) (err error) {
	sa.service.Log.Println("Got RPC admin command Unregister")
	sa.service.Unregister()
	return
}

type StopRequest struct {
	WaitForClients bool
}

type StopResponse struct {
}

func (sa *Admin) Stop(in StopRequest, out *StopResponse) (err error) {
	sa.service.Log.Println("Got RPC admin command Stop")

	// TODO: if in.WaitForClients is true, do it

	sa.service.Shutdown()
	return
}
