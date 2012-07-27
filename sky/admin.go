package main

import (
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/client"
	"github.com/bketelsen/skynet/rpc/bsonrpc"
	"net"
	"net/rpc"
	"os"
)

func doRegister(rpcClient *rpc.Client, log skynet.Logger) {
	var args skynet.RegisterParams
	var reply skynet.RegisterReturns
	err := rpcClient.Call("Admin.Register", args, &reply)
	if err != nil {
		log.Item(err)
	}
}

func doUnregister(rpcClient *rpc.Client, log skynet.Logger) {
	var args skynet.UnregisterParams
	var reply skynet.UnregisterReturns
	err := rpcClient.Call("Admin.Unregister", args, &reply)
	if err != nil {
		log.Item(err)
	}
}

func Register(q *client.Query) {
	doSomething(q, doRegister)
}

func Unregister(q *client.Query) {
	doSomething(q, doUnregister)
}

func doSomething(q *client.Query, do func(*rpc.Client, skynet.Logger)) {

	log := skynet.NewConsoleLogger(os.Stderr)
	for _, instance := range q.FindInstances() {
		conn, err := net.Dial("tcp", instance.Config.AdminAddr.String())
		if err != nil {
			log.Item(err)
			continue
		}
		rpcClient := bsonrpc.NewClient(conn)
		do(rpcClient, log)
		conn.Close()
	}
}
