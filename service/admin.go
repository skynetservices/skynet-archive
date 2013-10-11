package service

import (
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/log"
)

type Admin struct {
	service *Service
}

func (sa *Admin) Register(in skynet.RegisterRequest, out *skynet.RegisterResponse) (err error) {
	log.Println(log.TRACE, "Got RPC admin command Register")
	sa.service.Register()
	return
}

func (sa *Admin) Unregister(in skynet.UnregisterRequest, out *skynet.UnregisterResponse) (err error) {
	log.Println(log.TRACE, "Got RPC admin command Unregister")
	sa.service.Unregister()
	return
}

func (sa *Admin) Stop(in skynet.StopRequest, out *skynet.StopResponse) (err error) {
	log.Println(log.TRACE, "Got RPC admin command Stop")

	sa.service.Shutdown()
	return
}
