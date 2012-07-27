package skynet

import (
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

func (sa *ServiceAdmin) Listen(addr *BindAddr) {
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

type RegisterParams struct {
}

type RegisterReturns struct {
}

func (sa *Admin) Register(in RegisterParams, out *RegisterReturns) (err error) {
	sa.service.Log.Println("Got RPC admin command Register")
	sa.service.Register()
	return
}

type UnregisterParams struct {
}

type UnregisterReturns struct {
}

func (sa *Admin) Unregister(in UnregisterParams, out *UnregisterReturns) (err error) {
	sa.service.Log.Println("Got RPC admin command Unregister")
	sa.service.Unregister()
	return
}

type StopParams struct {
	WaitForClients bool
}

type StopReturns struct {
}

func (sa *Admin) Stop(in StopParams, out *StopReturns) (err error) {
	sa.service.Log.Println("Got RPC admin command Stop")

	sa.service.Shutdown()
	return
}
