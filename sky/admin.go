package main

import (
	"github.com/bketelsen/skynet/rpc/bsonrpc"
	"github.com/bketelsen/skynet/skylib"
	"net"
	"net/rpc"
	"os"
)

func doRegister(rpcClient *rpc.Client, log skylib.Logger) {
	var args skylib.RegisterParams
	var reply skylib.RegisterReturns
	err := rpcClient.Call("Admin.Register", args, &reply)
	if err != nil {
		log.Item(err)
	}
}

func doUnregister(rpcClient *rpc.Client, log skylib.Logger) {
	var args skylib.UnregisterParams
	var reply skylib.UnregisterReturns
	err := rpcClient.Call("Admin.Unregister", args, &reply)
	if err != nil {
		log.Item(err)
	}
}

func Register(q *skylib.Query) {
	doSomething(q, doRegister)
}

func Unregister(q *skylib.Query) {
	doSomething(q, doUnregister)
}

func doSomething(q *skylib.Query, do func(*rpc.Client, skylib.Logger)) {

	log := skylib.NewConsoleLogger(os.Stderr)
	results := *q.FindInstances()
	for _, result := range results {
		conn, err := net.Dial("tcp", result.Config.AdminAddr.String())
		if err != nil {
			log.Item(err)
			continue
		}
		rpcClient := bsonrpc.NewClient(conn)
		do(rpcClient, log)
		conn.Close()
	}
}
