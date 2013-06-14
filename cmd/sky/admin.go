package main

import (
	"fmt"
	"github.com/kballard/go-shellquote"
	"github.com/skynetservices/skynet2"
	"github.com/skynetservices/skynet2/daemon"
	"os"
	"text/template"
)

var deployTemplate = template.Must(template.New("").Parse(
	`Deployed service with UUID {{.UUID}}.
`))

func Start(criteria *skynet.Criteria, args []string) {
	if len(args) < 1 {
		fmt.Println("Please provide a service name 'sky start binaryName'")
		return
	}

	hosts, err := skynet.GetServiceManager().ListHosts(criteria)

	if err != nil {
		panic(err)
	}

	for _, host := range hosts {
		fmt.Println("Starting on host: " + host)
		d := daemon.GetDaemonForHost(Client, host)

		in := daemon.StartSubServiceRequest{
			BinaryName: args[0],
			Args:       shellquote.Join(args[1:]...),
		}
		out, err := d.StartSubService(in)

		if err != nil {
			fmt.Println("Returned Error: " + err.Error())
			return
		}

		deployTemplate.Execute(os.Stdout, out)
	}
}

var stopTemplate = template.Must(template.New("").Parse(
	`{{if .Ok}}Stopped service with UUID {{.UUID}}.
{{else}}Service with UUID {{.UUID}} is already stopped.
{{end}}`))

func Stop(criteria *skynet.Criteria) {
	instances, err := skynet.GetServiceManager().ListInstances(criteria)

	if err != nil {
		panic(err)
	}

	for _, instance := range filterDaemon(instances) {
		fmt.Println("Stopping: " + instance.UUID)
		d := daemon.GetDaemonForService(Client, &instance)

		in := daemon.StopSubServiceRequest{
			UUID: instance.UUID,
		}
		out, err := d.StopSubService(in)

		if err != nil {
			fmt.Println("Returned Error: " + err.Error())
			return
		}

		stopTemplate.Execute(os.Stdout, out)
	}
}

var restartTemplate = template.Must(template.New("").Parse(
	`{{if .Ok}}Restarted service with UUID {{.UUID}}.
{{else}}Service with UUID {{.UUID}} is not running.
{{end}}`))

func Restart(criteria *skynet.Criteria) {
	instances, err := skynet.GetServiceManager().ListInstances(criteria)

	if err != nil {
		panic(err)
	}

	for _, instance := range filterDaemon(instances) {
		fmt.Println("Restarting: " + instance.UUID)
		d := daemon.GetDaemonForService(Client, &instance)

		in := daemon.RestartSubServiceRequest{
			UUID: instance.UUID,
		}
		out, err := d.RestartSubService(in)

		if err != nil {
			fmt.Println("Returned Error: " + err.Error())
			return
		}

		restartTemplate.Execute(os.Stdout, out)
	}
}

var registerTemplate = template.Must(template.New("").Parse(
	`Registered service with UUID {{.UUID}}.
`))

func Register(criteria *skynet.Criteria) {
	instances, err := skynet.GetServiceManager().ListInstances(criteria)

	if err != nil {
		panic(err)
	}

	for _, instance := range filterDaemon(instances) {
		fmt.Println("Registering: " + instance.UUID)
		d := daemon.GetDaemonForService(Client, &instance)

		in := daemon.RegisterSubServiceRequest{
			UUID: instance.UUID,
		}
		out, err := d.RegisterSubService(in)

		if err != nil {
			fmt.Println("Returned Error: " + err.Error())
			return
		}

		registerTemplate.Execute(os.Stdout, out)
	}
}

var unregisterTemplate = template.Must(template.New("").Parse(
	`Unregistered service with UUID {{.UUID}}.
`))

func Unregister(criteria *skynet.Criteria) {
	instances, err := skynet.GetServiceManager().ListInstances(criteria)

	if err != nil {
		panic(err)
	}

	for _, instance := range filterDaemon(instances) {
		fmt.Println("Unregistering: " + instance.UUID)
		d := daemon.GetDaemonForService(Client, &instance)

		in := daemon.UnregisterSubServiceRequest{
			UUID: instance.UUID,
		}
		out, err := d.UnregisterSubService(in)

		if err != nil {
			fmt.Println("Returned Error: " + err.Error())
			return
		}

		unregisterTemplate.Execute(os.Stdout, out)
	}
}

var logLevelTemplate = template.Must(template.New("").Parse(
	`Set LogLevel to {{.Level}} for UUID {{.UUID}}.
`))

func SetLogLevel(criteria *skynet.Criteria, level string) {
	instances, err := skynet.GetServiceManager().ListInstances(criteria)

	if err != nil {
		panic(err)
	}

	for _, instance := range filterDaemon(instances) {
		fmt.Println("Setting LogLevel to " + level + " for: " + instance.UUID)
		d := daemon.GetDaemonForService(Client, &instance)

		in := daemon.SubServiceLogLevelRequest{
			UUID:  instance.UUID,
			Level: level,
		}
		out, err := d.SubServiceLogLevel(in)

		if err != nil {
			fmt.Println("Returned Error: " + err.Error())
			return
		}

		logLevelTemplate.Execute(os.Stdout, out)
	}
}

func filterDaemon(instances []skynet.ServiceInfo) []skynet.ServiceInfo {
	filteredInstances := make([]skynet.ServiceInfo, 0)

	for _, i := range instances {
		if i.Name != "SkynetDaemon" {
			filteredInstances = append(filteredInstances, i)
		}
	}

	return filteredInstances
}
