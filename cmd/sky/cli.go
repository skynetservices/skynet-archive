package main

import (
	"fmt"
	"github.com/bketelsen/skynet"
	"github.com/sbinet/liner"
	"strconv"
	"strings"
	"syscall"
)

var query *skynet.Query

/*
 * CLI Logic
 */

var SupportedCliCommands = []string{
	"deploy",
	"exit",
	"filters",
	"help",
	"host",
	"hosts",
	"instances",
	"port",
	"region",
	"regions",
	"register",
	"registered",
	"reset",
	"service",
	"services",
	"stop",
	"topology",
	"unregister",
	"version",
	"versions",
}

func tabCompleter(line string) []string {
	cmds := make([]string, 0)

	opts := make([]string, 0)

	if strings.HasPrefix(line, "reset") {
		filters := []string{
			"reset host",
			"reset port",
			"reset region",
			"reset registered",
			"reset service",
			"reset version",
		}

		for _, cmd := range filters {
			if strings.HasPrefix(cmd, line) {
				opts = append(opts, cmd)
			}
		}
	} else if strings.HasPrefix(line, "host") {
		cmds = make([]string, 0)

		for _, host := range query.FindHosts() {
			cmds = append(cmds, "host "+host)
		}
	} else if strings.HasPrefix(line, "region") {
		cmds = make([]string, 0)

		for _, region := range query.FindRegions() {
			cmds = append(cmds, "region "+region)
		}
	} else if strings.HasPrefix(line, "service") {
		cmds = make([]string, 0)

		for _, service := range query.FindServices() {
			cmds = append(cmds, "service "+service)
		}
	} else if strings.HasPrefix(line, "version") {
		cmds = make([]string, 0)

		for _, version := range query.FindServiceVersions() {
			cmds = append(cmds, "version "+version)
		}
	} else if strings.HasPrefix(line, "registered") {
		cmds = []string{"registered true", "registered false"}
	} else {
		cmds = SupportedCliCommands
	}

	for _, cmd := range cmds {
		if strings.HasPrefix(cmd, line) {
			opts = append(opts, cmd)
		}
	}

	return opts
}

func InteractiveShell() {
	term := liner.NewLiner()

	doozer := Doozer(config.DoozerConfig)

	fmt.Println("Skynet Interactive Shell")

	query = &skynet.Query{
		DoozerConn: doozer,
	}

	term.SetCompleter(tabCompleter)

	for {
		l, e := term.Prompt("> ")
		if e != nil {
			break
		}

		s := string(l)
		parts := strings.Split(s, " ")
		validCommand := true

		switch parts[0] {
		case "deploy":
			if len(parts) >= 2 {
				if confirm(term, "Service will be deployed to "+strconv.Itoa(len(query.FindHosts()))+" hosts") {
					Deploy(query, parts[1], parts[2:]...)
				}
			} else {
				fmt.Println("Usage: deploy <service path> <args>")
			}
		case "exit":
			term.Close()
			syscall.Exit(0)
		case "help", "h":
			InteractiveShellHelp()
		case "services":
			ListServices(query)
		case "hosts":
			ListHosts(query)
		case "regions":
			ListRegions(query)
		case "instances":
			ListInstances(query)
		case "versions":
			ListServiceVersions(query)
		case "topology":
			PrintTopology(query)

		case "service":
			if len(parts) >= 2 {
				query.Service = parts[1]
			}

			fmt.Printf("Service: %v\n", query.Service)

		case "host":
			if len(parts) >= 2 {
				query.Host = parts[1]
			}

			fmt.Printf("Host: %v\n", query.Host)

		case "port":
			if len(parts) >= 2 {
				query.Port = parts[1]
			}

			fmt.Printf("Host: %v\n", query.Host)

		case "version":
			if len(parts) >= 2 {
				query.Version = parts[1]
			}

			fmt.Printf("Version: %v\n", query.Version)

		case "region":
			if len(parts) >= 2 {
				query.Region = parts[1]
			}

			fmt.Printf("Region: %v\n", query.Region)

		case "register":
			if confirm(term, strconv.Itoa(len(filterDaemon(query.FindInstances())))+" instances will be registered") {
				Register(query)
			}
		case "unregister":
			if confirm(term, strconv.Itoa(len(filterDaemon(query.FindInstances())))+" instances will be unregistered") {
				Unregister(query)
			}
		case "stop":
			if confirm(term, strconv.Itoa(len(filterDaemon(query.FindInstances())))+" instances will be stopped") {
				Stop(query)
			}

		case "registered":
			if len(parts) >= 2 {
				var reg bool

				if parts[1] == "true" {
					reg = true
				} else {
					reg = false
				}

				query.Registered = &reg
			}

			registered := ""
			if query.Registered != nil {
				registered = strconv.FormatBool(*query.Registered)
			}

			fmt.Printf("Registered: %v\n", registered)

		case "reset":
			if len(parts) == 1 || parts[1] == "service" {
				query.Service = ""
			}

			if len(parts) == 1 || parts[1] == "version" {
				query.Version = ""
			}

			if len(parts) == 1 || parts[1] == "host" {
				query.Host = ""
			}

			if len(parts) == 1 || parts[1] == "port" {
				query.Port = ""
			}

			if len(parts) == 1 || parts[1] == "region" {
				query.Region = ""
			}

			if len(parts) == 1 || parts[1] == "registered" {
				query.Registered = nil
			}
		case "filters":
			registered := ""
			if query.Registered != nil {
				registered = strconv.FormatBool(*query.Registered)
			}

			fmt.Printf("Region: %v\nHost: %v\nService: %v\nVersion: %v\nRegistered: %v\n", query.Region, query.Host, query.Service, query.Version, registered)
		default:
			validCommand = false
			fmt.Println("Unknown Command - type 'help' for a list of commands")
		}

		if validCommand {
			term.AppendHistory(s)
		}
	}
}

func confirm(term *liner.State, msg string) bool {
	confirm, _ := term.Prompt(msg + ", Are you sure? (Y/N) > ")
	if confirm == "Y" || confirm == "y" {
		return true
	}

	return false
}

func filterDaemon(instances []*skynet.ServiceInfo) []*skynet.ServiceInfo {
	filteredInstances := make([]*skynet.ServiceInfo, 0)

	for _, i := range instances {
		if i.Config.Name != "SkynetDaemon" {
			filteredInstances = append(filteredInstances, i)
		}
	}

	return filteredInstances
}

func InteractiveShellHelp() {
	fmt.Print(`
Commands:
	deploy: Deploy new instances to cluster, will deploy to all hosts matching current filters (deploy <service path> <args>)
	hosts: List all hosts available that meet the specified criteria
	instances: List all instances available that meet the specified criteria
	regions: List all regions available that meet the specified criteria
	register: Registers all instances that match the current filters
	unregister: Unregisters all instances that match the current filters
	stop: Stops all instances that match the current filters
	services: List all services available that meet the specified criteria
	versions: List all services available that meet the specified criteria
	topology: Print detailed heirarchy of regions/hosts/services/versions/instances

Filters:
	filters - list current filters
	reset <filter> - reset all filters or specified filter
	region <region> - Set region filter, all commands will be scoped to this region until reset
	service <service> - Set service filter, all commands will be scoped to this service until reset
	version <version> - Set version filter, all commands will be scoped to this version until reset
	host <host> - Set host filter, all commands will be scoped to this host until reset
	port <port> - Set port filter, all commands will be scoped to this port until reset

`)
}
