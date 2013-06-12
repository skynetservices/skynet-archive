package main

/* TODO:
Implement Build/Deploy
Update filter prints so it prints comma separated list of filters for each type
Update help to make it clear you can keep adding multiple filters of a given type
Modify service filter to support versions (can use helper from sky.go)
*/

import (
	"fmt"
	"github.com/sbinet/liner"
	"github.com/skynetservices/skynet2"
	"strconv"
	"strings"
	"syscall"
)

var criteria = new(skynet.Criteria)

/*
* CLI Logic
 */

var SupportedCliCommands = []string{
	"exit",
	"filters",
	"help",
	"host",
	"hosts",
	"instances",
	"region",
	"regions",
	"registered",
	"reset",
	"service",
	"services",
	"version",
	"versions",
}

func tabCompleter(line string) []string {
	cmds := make([]string, 0)

	opts := make([]string, 0)

	if strings.HasPrefix(line, "reset") {
		filters := []string{
			"reset hosts",
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

		for _, host := range getHosts(criteria) {
			cmds = append(cmds, "host "+host)
		}
	} else if strings.HasPrefix(line, "region") {
		cmds = make([]string, 0)

		for _, region := range getRegions(criteria) {
			cmds = append(cmds, "region "+region)
		}
	} else if strings.HasPrefix(line, "service") {
		cmds = make([]string, 0)

		for _, service := range getServices(criteria) {
			cmds = append(cmds, "service "+service)
		}
	} else if strings.HasPrefix(line, "version") {
		cmds = make([]string, 0)

		for _, version := range getVersions(criteria) {
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

	fmt.Println("Skynet Interactive Shell")

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
		case "exit", "quit":
			term.Close()
			syscall.Exit(0)
		case "help", "h":
			InteractiveShellHelp()
		case "services":
			ListServices(criteria)
		case "hosts":
			ListHosts(criteria)
		case "regions":
			ListRegions(criteria)
		case "instances":
			ListInstances(criteria)
		case "versions":
			ListVersions(criteria)

		case "service":
			if len(parts) >= 2 {
				criteria.Services = []skynet.ServiceCriteria{
					skynet.ServiceCriteria{Name: parts[1]},
				}
			}

			fmt.Printf("Service: %v\n", criteria.Services)

		case "host":
			if len(parts) >= 2 {
				criteria.Hosts = append(criteria.Hosts, parts[1])
			}

			fmt.Printf("Host: %v\n", criteria.Hosts)

		case "region":
			if len(parts) >= 2 {
				criteria.Regions = append(criteria.Regions, parts[1])
			}

			fmt.Printf("Region: %v\n", criteria.Regions)

		case "registered":
			if len(parts) >= 2 {
				var reg bool

				if parts[1] == "true" {
					reg = true
				} else {
					reg = false
				}

				criteria.Registered = &reg
			}

			registered := ""
			if criteria.Registered != nil {
				registered = strconv.FormatBool(*criteria.Registered)
			}

			fmt.Printf("Registered: %v\n", registered)

		case "reset":
			if len(parts) == 1 || parts[1] == "service" {
				criteria.Services = []skynet.ServiceCriteria{}
			}

			if len(parts) == 1 || parts[1] == "host" {
				criteria.Hosts = []string{}
			}

			if len(parts) == 1 || parts[1] == "region" {
				criteria.Regions = []string{}
			}

			if len(parts) == 1 || parts[1] == "registered" {
				criteria.Registered = nil
			}
		case "filters":
			registered := ""
			if criteria.Registered != nil {
				registered = strconv.FormatBool(*criteria.Registered)
			}

			fmt.Printf("Region: %v\nHost: %v\nService: %v\nRegistered: %v\n", criteria.Regions, criteria.Hosts, criteria.Services, registered)
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

func filterDaemon(instances []skynet.ServiceInfo) []skynet.ServiceInfo {
	filteredInstances := make([]skynet.ServiceInfo, 0)

	for _, i := range instances {
		if i.Name != "SkynetDaemon" {
			filteredInstances = append(filteredInstances, i)
		}
	}

	return filteredInstances
}

func InteractiveShellHelp() {
	fmt.Print(`
  Commands:
  hosts: List all hosts available that meet the specified criteria
  instances: List all instances available that meet the specified criteria
  regions: List all regions available that meet the specified criteria
  services: List all services available that meet the specified criteria
  versions: List all services available that meet the specified criteria

  Filters:
  filters - list current filters
  reset <filter> - reset all filters or specified filter
  region <region> - Set region filter, all commands will be scoped to this region until reset
  service <service> - Set service filter, all commands will be scoped to this service until reset
  version <version> - Set version filter, all commands will be scoped to this version until reset
  host <host> - Set host filter, all commands will be scoped to this host until reset

  `)
}
