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

func (sa *ServiceAdmin) Listen(addr *skynet.BindAddr, bindChan chan bool) {
	listener, err := addr.Listen()
	if err != nil {
		panic(err)
	}

	bindChan <- true
	sa.service.Log.Item(skynet.AdminListening{sa.service.Config})

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

func (sa *Admin) Register(in skynet.RegisterRequest, out *skynet.RegisterResponse) (err error) {
	sa.service.Log.Println("Got RPC admin command Register")
	sa.service.Register()
	return
}

func (sa *Admin) Unregister(in skynet.UnregisterRequest, out *skynet.UnregisterResponse) (err error) {
	sa.service.Log.Println("Got RPC admin command Unregister")
	sa.service.Unregister()
	return
}

func (sa *Admin) Stop(in skynet.StopRequest, out *skynet.StopResponse) (err error) {
	sa.service.Log.Item("Got RPC admin command Stop")

	// TODO: if in.WaitForClients is true, do it

	sa.service.Shutdown()
	return
}
