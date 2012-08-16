package main

import (
	"fmt"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/client"
	"os"
	"text/template"
)

// Remote() uses the SkynetDaemon service to remotely manage services.
func Remote(q *client.Query, args []string) {
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

func getDaemonServiceClient(q *client.Query) (c *client.Client, service *client.ServiceClient) {
	config, _ := skynet.GetClientConfigFromFlags(os.Args...)

	config.Log = skynet.NewConsoleLogger(os.Stderr)

	c = client.NewClient(config)

	registered := true
	query := &client.Query{
		DoozerConn: c.DoozerConn,
		Service:    "SkynetDaemon",
		//Host:       "127.0.0.1",
		Registered: &registered,
	}
	service = c.GetServiceFromQuery(query)
	return
}

var listTemplate = template.Must(template.New("").Parse(
	`Services deployed on this daemon:{{range .Services}}
UUID: {{.UUID}}
 {{.ServicePath}}
 {{.Args}}
 {{if .Running}}RUNNING{{else}}HALTED{{end}}
{{end}}
`))

func remoteList(q *client.Query) {
	_, service := getDaemonServiceClient(q)

	// This on the other hand will fail if it can't find a service to connect to
	var response ListSubServicesOut
	err := service.Send(nil, "ListSubServices", ListSubServicesIn{}, &response)

	if err != nil {
		fmt.Println(err)
		return
	}

	listTemplate.Execute(os.Stdout, response)
}

func remoteDeploy(q *client.Query, servicePath string, serviceArgs []string) {

}

func remoteStart(q *client.Query, uuid string) {
	_, service := getDaemonServiceClient(q)

	// This on the other hand will fail if it can't find a service to connect to
	var in = StartSubServiceIn{UUID: uuid}
	var out StartSubServiceOut
	err := service.Send(nil, "StartSubService", in, &out)

	if err != nil {
		fmt.Println(err)
		return
	}

	if out.Ok {
		fmt.Printf("Started subservice with UUID %s.\n", out.UUID)
	} else {
		fmt.Printf("Subservice with UUID %s is already running.\n", out.UUID)
	}
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
