package main

import (
	"fmt"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/client"
	"github.com/bketelsen/skynet/daemon"
	"github.com/bketelsen/skynet/rpc/bsonrpc"
	"github.com/bketelsen/skynet/service"
	"github.com/kballard/go-shellquote"
	"net"
	"net/rpc"
	"os"
	"strings"
	"text/template"
)

func doRegister(rpcClient *rpc.Client, log skynet.Logger) {
	var args service.RegisterParams
	var reply service.RegisterReturns
	err := rpcClient.Call("Admin.Register", args, &reply)
	if err != nil {
		log.Item(err)
	}
}

func doUnregister(rpcClient *rpc.Client, log skynet.Logger) {
	var args service.UnregisterParams
	var reply service.UnregisterReturns
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

func getDaemonServiceClientForHost(dc *skynet.DoozerConfig, host string) *client.ServiceClient {
	config := &skynet.ClientConfig{
		DoozerConfig: dc,
	}

	c := client.NewClient(config)
	registered := true
	query := &client.Query{
		DoozerConn: c.DoozerConn,
		Service:    "SkynetDaemon",
		Host:       host,
		Registered: &registered,
	}

	s := c.GetServiceFromQuery(query)
	return s
}

var deployTemplate = template.Must(template.New("").Parse(
	`Deployed service with UUID {{.UUID}}.
`))

// TODO: this should be smarter about which hosts it deploys to
func Deploy(q *client.Query, path string, args ...string) {
	fmt.Println("deploying " + path + " " + strings.Join(args, ""))

	for _, host := range q.FindHosts() {
		if _, ok := serviceClients[host]; !ok {
			serviceClients[host] = getDaemonServiceClientForHost(q.DoozerConn.Config, host)
		}

		s := serviceClients[host]

		in := daemon.DeployRequest{
			ServicePath: path,
			Args:        shellquote.Join(args...),
		}
		var out daemon.DeployResponse

		err := s.Send(nil, "Deploy", in, &out)

		if err != nil {
			fmt.Println(err)
			return
		}

		deployTemplate.Execute(os.Stdout, out)
	}
}

var stopTemplate = template.Must(template.New("").Parse(
	`{{if .Ok}}Stopped service with UUID {{.UUID}}.
{{else}}Service with UUID {{.UUID}} is already stopped.
{{end}}`))

var serviceClients = make(map[string]*client.ServiceClient)

func Stop(q *client.Query) {
	for _, instance := range q.FindInstances() {
		if _, ok := serviceClients[instance.Config.ServiceAddr.IPAddress]; !ok {
			serviceClients[instance.Config.ServiceAddr.IPAddress] = getDaemonServiceClientForHost(q.DoozerConn.Config, instance.Config.ServiceAddr.IPAddress)
		}

		s := serviceClients[instance.Config.ServiceAddr.IPAddress]

		in := daemon.StopSubServiceRequest{UUID: instance.Config.UUID}
		out := daemon.StopSubServiceResponse{}

		err := s.Send(nil, "StopSubService", in, &out)

		if err != nil {
			fmt.Println(err)
		} else {
			stopTemplate.Execute(os.Stdout, out)
		}
	}
}
