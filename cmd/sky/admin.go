package main

import (
	"fmt"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/client"
	"github.com/bketelsen/skynet/daemon"
	"github.com/bketelsen/skynet/service"
	"github.com/kballard/go-shellquote"
	"os"
	"strings"
	"text/template"
)

func Register(q *client.Query) {
	instances := q.FindInstances()
	for _, instance := range instances {
		cladmin := client.Admin{
			Instance: instance,
		}
		_, err := cladmin.Register(service.RegisterRequest{})
		if err != nil {
			config.Log.Item(err)
		}
	}
}

func Unregister(q *client.Query) {
	instances := q.FindInstances()
	for _, instance := range instances {
		cladmin := client.Admin{
			Instance: instance,
		}
		_, err := cladmin.Unregister(service.UnregisterRequest{})
		if err != nil {
			config.Log.Item(err)
		}
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
	cl := client.NewClient(&config)

	fmt.Println("deploying " + path + " " + strings.Join(args, ""))

	for _, host := range q.FindHosts() {
		cdaemon := daemon.GetDaemonForHost(cl, host)

		in := daemon.DeployRequest{
			ServicePath: path,
			Args:        shellquote.Join(args...),
		}
		out, err := cdaemon.Deploy(in)

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

func Stop(q *client.Query) {
	cl := client.NewClient(&config)

	for _, instance := range q.FindInstances() {
		cdaemon := daemon.GetDaemonForService(cl, instance)

		in := daemon.StopSubServiceRequest{UUID: instance.Config.UUID}
		out, err := cdaemon.StopSubService(in)

		if err != nil {
			if strings.HasPrefix(err.Error(), "No such service UUID") {
				// no daemon on the service's machine, shut it down directly
				AdminStop(q)
			} else {
				fmt.Println(err)
			}
		} else {
			stopTemplate.Execute(os.Stdout, out)
		}
	}
}

func AdminStop(q *client.Query) {
	instances := q.FindInstances()
	for _, instance := range instances {
		cladmin := client.Admin{
			Instance: instance,
		}
		_, err := cladmin.Stop(service.StopRequest{
			WaitForClients: true,
		})
		if err != nil {
			config.Log.Item(err)
		}
	}
}
