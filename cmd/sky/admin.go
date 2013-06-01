package main

import (
	"fmt"
	"github.com/skynetservices/go-shellquote"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/client"
	"github.com/skynetservices/skynet/daemon"
	"github.com/skynetservices/skynet/log"
	"os"
	"strings"
	"text/template"
)

func Register(q *skynet.Query) {
	instances := filterDaemon(q.FindInstances())
	for _, instance := range instances {
		cladmin := client.Admin{
			Instance: instance,
		}
		_, err := cladmin.Register(skynet.RegisterRequest{})
		if err != nil {
			log.Println(log.ERROR, err.Error())
		}
	}
}

func Unregister(q *skynet.Query) {
	instances := filterDaemon(q.FindInstances())
	for _, instance := range instances {
		cladmin := client.Admin{
			Instance: instance,
		}
		_, err := cladmin.Unregister(skynet.UnregisterRequest{})
		if err != nil {
			log.Println(log.ERROR, err.Error())
		}
	}
}

func getDaemonServiceClientForHost(dc *skynet.DoozerConfig, host string) *client.ServiceClient {
	config := &skynet.ClientConfig{
		DoozerConfig:             dc,
		MaxConnectionsToInstance: 10,
	}

	c := client.NewClient(config)
	registered := true
	query := &skynet.Query{
		DoozerConn: c.DoozerConn,
		Service:    "SkynetDaemon",
		Host:       host,
		Registered: &registered,
	}

	s := c.GetServiceFromQuery(query)
	return s
}

var startTemplate = template.Must(template.New("").Parse(
	`Started service with UUID {{.UUID}}.
`))

// TODO: this should be smarter about which hosts it starts on
func Start(q *skynet.Query, path string, args ...string) {
	cl := client.NewClient(config)

	fmt.Println("starting " + path + " " + strings.Join(args, ""))

	for _, host := range q.FindHosts() {
		cdaemon := daemon.GetDaemonForHost(cl, host)

		in := daemon.StartRequest{
			ServicePath: path,
			Args:        shellquote.Join(args...),
		}
		out, err := cdaemon.Start(in)

		if err != nil {
			fmt.Println(err)
			return
		}

		startTemplate.Execute(os.Stdout, out)
	}
}

var stopTemplate = template.Must(template.New("").Parse(
	`{{if .Ok}}Stopped service with UUID {{.UUID}}.
{{else}}Service with UUID {{.UUID}} is already stopped.
{{end}}`))

func Stop(q *skynet.Query) {
	cl := client.NewClient(config)

	for _, instance := range filterDaemon(q.FindInstances()) {
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

var restartTemplate = template.Must(template.New("").Parse(
	`{{if .Ok}}Restarted service with UUID {{.UUID}}.
{{else}}Service with UUID {{.UUID}} is not running.
{{end}}`))

func Restart(q *skynet.Query) {
	cl := client.NewClient(config)

	for _, instance := range filterDaemon(q.FindInstances()) {
		cdaemon := daemon.GetDaemonForService(cl, instance)

		in := daemon.RestartSubServiceRequest{UUID: instance.Config.UUID}
		out, err := cdaemon.RestartSubService(in)

		if err != nil {
			if strings.HasPrefix(err.Error(), "No such service UUID") {
				// Commented out for now, we need to determine if we want to try to restart an unmanaged instance, and support it
				// no daemon on the service's machine, shut it down directly
				//AdminStop(q)
			} else {
				fmt.Println(err)
			}
		} else {
			restartTemplate.Execute(os.Stdout, out)
		}
	}

}

func AdminStop(q *skynet.Query) {
	instances := q.FindInstances()
	for _, instance := range instances {
		cladmin := client.Admin{
			Instance: instance,
		}
		_, err := cladmin.Stop(skynet.StopRequest{
			WaitForClients: true,
		})
		if err != nil {
			log.Println(log.ERROR, err.Error())
		}
	}
}
