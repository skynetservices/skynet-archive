package main

import (
	"fmt"
	"github.com/kballard/go-shellquote"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/daemon"
	"github.com/skynetservices/skynet/log"
	"os"
	"sync"
	"text/template"
)

var startTemplate = template.Must(template.New("").Parse(
	`Started service with UUID {{.UUID}}.
`))

func Start(criteria *skynet.Criteria, args []string) {
	if len(args) < 1 {
		fmt.Println("Please provide a service name 'sky start binaryName'")
		return
	}

	hosts, err := skynet.GetServiceManager().ListHosts(criteria)

	if err != nil {
		log.Fatal(err)
	}

	var wait sync.WaitGroup

	for _, host := range hosts {
		wait.Add(1)
		go func(host string) {
			fmt.Println("Starting on host: " + host)
			d := daemon.GetDaemonForHost(host)

			in := daemon.StartSubServiceRequest{
				BinaryName: args[0],
				Args:       shellquote.Join(args[1:]...),
				// TODO: maybe an optional flag to change this?
				Registered: true,
			}
			out, err := d.StartSubService(in)

			if err != nil {
				fmt.Println("Returned Error: " + err.Error())
				wait.Done()
				return
			}

			startTemplate.Execute(os.Stdout, out)
			wait.Done()
		}(host)
	}

	wait.Wait()
}

var stopTemplate = template.Must(template.New("").Parse(
	`{{if .Ok}}Stopped service with UUID {{.UUID}}.
{{else}}Service with UUID {{.UUID}} is already stopped.
{{end}}`))

func Stop(criteria *skynet.Criteria) {
	instances, err := skynet.GetServiceManager().ListInstances(criteria)

	if err != nil {
		log.Fatal(err)
	}

	var wait sync.WaitGroup

	for _, instance := range filterDaemon(instances) {
		wait.Add(1)
		go func(instance skynet.ServiceInfo) {
			fmt.Println("Stopping: " + instance.UUID)
			d := daemon.GetDaemonForService(&instance)

			in := daemon.StopSubServiceRequest{
				UUID: instance.UUID,
			}
			out, err := d.StopSubService(in)

			if err != nil {
				fmt.Println("Returned Error: " + err.Error())
				wait.Done()
				return
			}

			stopTemplate.Execute(os.Stdout, out)
			wait.Done()
		}(instance)
	}

	wait.Wait()
}

var restartTemplate = template.Must(template.New("").Parse(
	`{{if .Ok}}Restarted service with UUID {{.UUID}}.
{{else}}Service with UUID {{.UUID}} is not running.
{{end}}`))

func Restart(criteria *skynet.Criteria) {
	instances, err := skynet.GetServiceManager().ListInstances(criteria)

	if err != nil {
		log.Fatal(err)
	}

	var wait sync.WaitGroup

	for _, instance := range filterDaemon(instances) {
		wait.Add(1)
		go func(instance skynet.ServiceInfo) {
			fmt.Println("Restarting: " + instance.UUID)
			d := daemon.GetDaemonForService(&instance)

			in := daemon.RestartSubServiceRequest{
				UUID: instance.UUID,
			}
			out, err := d.RestartSubService(in)

			if err != nil {
				fmt.Println("Returned Error: " + err.Error())
				wait.Done()
				return
			}

			restartTemplate.Execute(os.Stdout, out)
			wait.Done()
		}(instance)
	}

	wait.Wait()
}

var registerTemplate = template.Must(template.New("").Parse(
	`Registered service with UUID {{.UUID}}.
`))

func Register(criteria *skynet.Criteria) {
	instances, err := skynet.GetServiceManager().ListInstances(criteria)

	if err != nil {
		log.Fatal(err)
	}

	var wait sync.WaitGroup

	for _, instance := range filterDaemon(instances) {
		wait.Add(1)
		go func(instance skynet.ServiceInfo) {
			fmt.Println("Registering: " + instance.UUID)
			d := daemon.GetDaemonForService(&instance)

			in := daemon.RegisterSubServiceRequest{
				UUID: instance.UUID,
			}
			out, err := d.RegisterSubService(in)

			if err != nil {
				fmt.Println("Returned Error: " + err.Error())
				wait.Done()
				return
			}

			registerTemplate.Execute(os.Stdout, out)
			wait.Done()
		}(instance)
	}

	wait.Wait()
}

var unregisterTemplate = template.Must(template.New("").Parse(
	`Unregistered service with UUID {{.UUID}}.
`))

func Unregister(criteria *skynet.Criteria) {
	instances, err := skynet.GetServiceManager().ListInstances(criteria)

	if err != nil {
		log.Fatal(err)
	}

	var wait sync.WaitGroup

	for _, instance := range filterDaemon(instances) {
		wait.Add(1)
		go func(instance skynet.ServiceInfo) {
			fmt.Println("Unregistering: " + instance.UUID)
			d := daemon.GetDaemonForService(&instance)

			in := daemon.UnregisterSubServiceRequest{
				UUID: instance.UUID,
			}
			out, err := d.UnregisterSubService(in)

			if err != nil {
				fmt.Println("Returned Error: " + err.Error())
				wait.Done()
				return
			}

			unregisterTemplate.Execute(os.Stdout, out)
			wait.Done()
		}(instance)
	}

	wait.Wait()
}

var logLevelTemplate = template.Must(template.New("").Parse(
	`Set LogLevel to {{.Level}} for UUID {{.UUID}}.
`))

func SetLogLevel(criteria *skynet.Criteria, level string) {
	instances, err := skynet.GetServiceManager().ListInstances(criteria)

	if err != nil {
		log.Fatal(err)
	}

	var wait sync.WaitGroup

	for _, instance := range filterDaemon(instances) {
		wait.Add(1)
		go func(instance skynet.ServiceInfo) {
			fmt.Println("Setting LogLevel to " + level + " for: " + instance.UUID)
			d := daemon.GetDaemonForService(&instance)

			in := daemon.SubServiceLogLevelRequest{
				UUID:  instance.UUID,
				Level: level,
			}
			out, err := d.SubServiceLogLevel(in)

			if err != nil {
				fmt.Println("Returned Error: " + err.Error())
				wait.Done()
				return
			}

			logLevelTemplate.Execute(os.Stdout, out)
			wait.Done()
		}(instance)
	}

	wait.Wait()
}

func SetDaemonLogLevel(criteria *skynet.Criteria, level string) {
	hosts, err := skynet.GetServiceManager().ListHosts(criteria)

	if err != nil {
		log.Fatal(err)
	}

	var wait sync.WaitGroup

	for _, host := range hosts {
		wait.Add(1)
		go func(host string) {
			d := daemon.GetDaemonForHost(host)

			in := daemon.LogLevelRequest{
				Level: level,
			}
			out, err := d.LogLevel(in)

			if err != nil {
				fmt.Println("Returned Error: " + err.Error())
				wait.Done()
				return
			}

			if out.Ok {
				fmt.Printf("Set daemon log level to %v on host: %v\n", level, host)
			} else {
				fmt.Printf("Failed to set daemon log level to %v on host: %v\n", level, host)
			}

			wait.Done()
		}(host)
	}

	wait.Wait()
}

func StopDaemon(criteria *skynet.Criteria) {
	hosts, err := skynet.GetServiceManager().ListHosts(criteria)

	if err != nil {
		log.Fatal(err)
	}

	var wait sync.WaitGroup

	for _, host := range hosts {
		wait.Add(1)
		go func(host string) {
			d := daemon.GetDaemonForHost(host)

			in := daemon.StopRequest{}
			out, err := d.Stop(in)

			if err != nil {
				fmt.Println("Returned Error: " + err.Error())
				wait.Done()
				return
			}

			if out.Ok {
				fmt.Printf("Daemon stopped on host: %v\n", host)
			} else {
				fmt.Printf("Failed to stop daemon on host: %v\n", host)
			}

			wait.Done()
		}(host)
	}

	wait.Wait()
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
