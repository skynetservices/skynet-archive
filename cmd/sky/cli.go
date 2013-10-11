package main

/* TODO:
Implement Build/Deploy
*/

import (
	"fmt"
	"github.com/sbinet/liner"
	"github.com/skynetservices/skynet"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

var criteria = new(skynet.Criteria)
var configFile = "./build.cfg"

/*
* CLI Logic
 */

var SupportedCliCommands = []string{
	"exit",
	"quit",
	"filters",
	"config",
	"deploy",
	"build",
	"help",
	"host",
	"hosts",
	"instance",
	"instances",
	"region",
	"regions",
	"registered",
	"reset",
	"service",
	"services",
	"version",
	"versions",
	"start",
	"stop",
	"restart",
	"register",
	"unregister",
	"log",
	"daemon",
	"daemon",
}

var serviceRegex = regexp.MustCompile("service ([^:]+):")

func tabCompleter(line string) []string {
	cmds := make([]string, 0)

	opts := make([]string, 0)

	if strings.HasPrefix(line, "reset") {
		filters := []string{
			"reset hosts",
			"reset instance",
			"reset region",
			"reset registered",
			"reset service",
			"reset version",
			"reset config",
		}

		for _, cmd := range filters {
			if strings.HasPrefix(cmd, line) {
				opts = append(opts, cmd)
			}
		}
	} else if strings.HasPrefix(line, "host") {
		cmds = make([]string, 0)

		for _, host := range getHosts(&skynet.Criteria{}) {
			if !exists(criteria.Hosts, host) {
				cmds = append(cmds, "host "+host)
			}
		}
	} else if strings.HasPrefix(line, "instance") {
		cmds = make([]string, 0)

		for _, instance := range getInstances(&skynet.Criteria{}) {
			if !exists(criteria.Instances, instance.UUID) {
				cmds = append(cmds, "instance "+instance.UUID)
			}
		}
	} else if strings.HasPrefix(line, "region") {
		cmds = make([]string, 0)

		for _, region := range getRegions(&skynet.Criteria{}) {
			if !exists(criteria.Regions, region) {
				cmds = append(cmds, "region "+region)
			}
		}
	} else if serviceRegex.MatchString(line) {
		cmds = make([]string, 0)
		matches := serviceRegex.FindAllStringSubmatch(line, -1)
		name := matches[0][1]

		c := new(skynet.Criteria)
		c.Services = []skynet.ServiceCriteria{skynet.ServiceCriteria{Name: name}}

		for _, version := range getVersions(c) {
			cmds = append(cmds, "service "+name+":"+version)
		}
	} else if strings.HasPrefix(line, "service") {
		cmds = make([]string, 0)

		for _, service := range getServices(&skynet.Criteria{}) {
			cmds = append(cmds, "service "+service)
		}
	} else if strings.HasPrefix(line, "registered") {
		cmds = []string{"registered true", "registered false"}
	} else if strings.HasPrefix(line, "log") {
		cmds = append([]string{"log DEBUG", "log TRACE", "log INFO", "log WARN", "log FATAL", "log PANIC"})
	} else if strings.HasPrefix(line, "daemon log") {
		cmds = append([]string{"daemon log DEBUG", "daemon log TRACE", "daemon log INFO", "daemon log WARN", "daemon log FATAL", "log PANIC"})
	} else if strings.HasPrefix(line, "daemon") {
		cmds = append([]string{"daemon log", "daemon stop"})
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
		case "build":
			Build(configFile)
		case "deploy":
			Deploy(configFile, criteria)
		case "start":
			Start(criteria, parts[1:])
		case "stop":
			Stop(criteria)
		case "restart":
			Restart(criteria)
		case "register":
			Register(criteria)
		case "unregister":
			Unregister(criteria)
		case "log":
			SetLogLevel(criteria, parts[1])
		case "daemon":
			if len(parts) >= 2 {
				switch parts[1] {
				case "log":
					if len(parts) >= 3 {
						SetDaemonLogLevel(criteria, parts[2])
					} else {
						fmt.Println("Must supply a log level")
					}
				case "stop":
					StopDaemon(criteria)
				}
			} else {
				fmt.Println("Supported subcommands for daemon are log, and stop")
			}

		case "config":
			if len(parts) >= 2 {
				configFile = parts[1]
			}

			fmt.Printf("Config: %s\n", configFile)

		case "service":
			if len(parts) >= 2 {
				criteria.AddService(serviceCriteriaFromString(parts[1]))
			}

			fmt.Printf("Services: %v\n", serviceCriteriaToString(criteria.Services))

		case "instance":
			if len(parts) >= 2 {
				criteria.AddInstance(parts[1])
			}

			fmt.Printf("Instances: %v\n", strings.Join(criteria.Instances, ", "))

		case "host":
			if len(parts) >= 2 {
				criteria.AddHost(parts[1])
			}

			fmt.Printf("Host: %v\n", strings.Join(criteria.Hosts, ", "))

		case "region":
			if len(parts) >= 2 {
				criteria.AddRegion(parts[1])
			}

			fmt.Printf("Region: %v\n", strings.Join(criteria.Regions, ", "))

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
			if len(parts) == 1 || parts[1] == "config" {
				configFile = "./build.cfg"
			}

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

			fmt.Printf("Region: %v\nHost: %v\nService: %v\nRegistered: %v\nInstances: %v\n", strings.Join(criteria.Regions, ", "), strings.Join(criteria.Hosts, ", "), serviceCriteriaToString(criteria.Services), registered, strings.Join(criteria.Instances, ", "))
		case "":
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

func InteractiveShellHelp() {
	fmt.Print(`
  Commands:
  hosts: List all hosts available that meet the specified criteria
  instances: List all instances available that meet the specified criteria
  regions: List all regions available that meet the specified criteria
  services: List all services available that meet the specified criteria
  versions: List all services available that meet the specified criteria
  config: Set config file for Build/Deploy (defaults to ./build.cfg)
  log: Set log level of services that meet the specified criteria log <level>, options are DEBUG, TRACE, INFO, WARN, FATAL, PANIC
  daemon log: Set log level of daemons that meet the specified criteria daemon log <level>, options are DEBUG, TRACE, INFO, WARN, FATAL, PANIC
  daemon stop: Stop daemons that match the specified criteria

  Filters:
  filters - list current filters
  reset <filter> - reset all filters or specified filter
  region <region> - Add a region to filter, all commands will be scoped to these regions until reset
  service <service> - Add a service to filter, all commands will be scoped to these services until reset
  host <host> - Add host to filter, all commands will be scoped to these hosts until reset
  instance <uuid> - Add an instance to filter, all commands will be scoped to this instance until reset

  `)
}

func exists(haystack []string, needle string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}

	return false
}

func serviceCriteriaToString(sc []skynet.ServiceCriteria) string {
	s := ""

	for _, v := range sc {
		s = s + v.String() + ", "
	}

	return s[:len(s)-2]
}
