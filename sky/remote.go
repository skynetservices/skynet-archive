package main

import (
	"fmt"
	"github.com/bketelsen/skynet/skylib"
	"os"
)

// Remote() uses the SkynetDaemon service to remotely manage services.
func Remote(q *skylib.Query, args []string) {
	if len(args) == 0 {
		remoteHelp()
		return
	}
	switch args[0] {
	case "list":
		remoteList(q)
	case "deploy":
		if len(args) < 2 {
			fmt.Printf("Must specify a service path")
			remoteHelp()
			return
		}
		servicePath := args[1]
		serviceArgs := args[2:]
		remoteDeploy(q, servicePath, serviceArgs)
	case "start":
		if len(args) != 2 {
			fmt.Printf("Must specify a service UUID")
			remoteHelp()
			return
		}
		uuid := args[1]
		remoteStart(q, uuid)
	case "help":
		remoteHelp()
	default:
		fmt.Printf("Unknown command %q", args[0])
		remoteHelp()
	}
	return
}

func getDaemonServiceClient(q *skylib.Query) (client *skylib.Client, service *skylib.ServiceClient) {
	config := &skylib.ClientConfig{
		DoozerConfig: &skylib.DoozerConfig{
			Uri:          "127.0.0.1:8046",
			AutoDiscover: true,
		},
	}

	config.Log = skylib.NewConsoleLogger(os.Stderr)

	client = skylib.NewClient(config)

	registered := true
	query := &skylib.Query{
		DoozerConn: client.DoozerConn,
		Service:    "SkynetDaemon",
		Host:       "127.0.0.1",
		Registered: &registered,
	}
	service = client.GetServiceFromQuery(query)
	return
}

func remoteList(q *skylib.Query) {
	_, service := getDaemonServiceClient(q)

	// This on the other hand will fail if it can't find a service to connect to
	var x struct{}
	ret := map[string]interface{}{}
	err := service.Send("ListSubServices", x, ret)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(ret)
}

func remoteDeploy(q *skylib.Query, servicePath string, serviceArgs []string) {

}

func remoteStart(q *skylib.Query, uuid string) {
	_, service := getDaemonServiceClient(q)

	

	// This on the other hand will fail if it can't find a service to connect to
	var in = M{"uuid": uuid}
	var out = M{}
	err := service.Send("StartSubService", in, out)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(out)
}

func remoteHelp() {
	fmt.Println(`remote commands:
	help
		- Print this help text.
	list
		- List all services currently being run by this daemon, with their uuids.
	deploy [service path] [arguments]
		- Deploy the service specified by the path, launched with the given arguments.
		  The uuid of the service will be printed.
	start [uuid]
		- Start the service assined to the given uuid.
	stop [uuid]
		- Stop the service assined to the given uuid.
	restart [uuid]
		- Restart the service assined to the given uuid.
	register [uuid]
		- Register the service assined to the given uuid.
	deregister [uuid]
		- Deregister the service assined to the given uuid.
	
`)
}
