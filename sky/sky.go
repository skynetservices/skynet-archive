package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/client"
	"github.com/bketelsen/skynet/service"
	"os"
	"strconv"
	"strings"
	"syscall"
)

var (
	DoozerHost      *string = flag.String("doozer", skynet.GetDefaultEnvVar("DZHOST", "127.0.0.1:8046"), "initial doozer instance to connect to")
	VersionFlag     *string = flag.String("version", "", "service version")
	ServiceNameFlag *string = flag.String("service", "", "service name")
	HostFlag        *string = flag.String("host", "", "host")
	PortFlag        *string = flag.String("port", "", "port")
	RegionFlag      *string = flag.String("region", "", "region")
	RegisteredFlag  *string = flag.String("registered", "", "registered")
)

var DC *skynet.DoozerConnection

func main() {
	flag.Parse()

	query := &client.Query{
		DoozerConn: Doozer(),
		Service:    *ServiceNameFlag,
		Version:    *VersionFlag,
		Host:       *HostFlag,
		Region:     *RegionFlag,
		Port:       *PortFlag,
	}

	switch flag.Arg(0) {
	case "help", "h":
		CommandLineHelp()
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
	case "daemon":
		Daemon(query, flag.Args()[1:])
	case "remote":
		Remote(query, flag.Args()[1:])
	case "register":
		Register(query)
	case "unregister":
		Unregister(query)
	case "cli":
		InteractiveShell()

	default:
		CommandLineHelp()
	}
}

func Doozer() *skynet.DoozerConnection {
	if DC == nil {
		DC = Connect()
	}

	return DC
}

func Connect() *skynet.DoozerConnection {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Failed to connect to Doozer")
			os.Exit(1)
		}
	}()

	// TODO: This needs to come from command line, or environment variable
	conn := skynet.NewDoozerConnection(*DoozerHost, "", false, nil) // nil as the last param will default to a Stdout logger
	conn.Connect()

	return conn
}

func ListInstances(q *client.Query) {
	var regFlag *bool

	if *RegisteredFlag == "true" {
		b := true
		regFlag = &b
	} else if *RegisteredFlag == "false" {
		b := false
		regFlag = &b
	}

	q.Registered = regFlag

	results := q.FindInstances()

	for _, instance := range results {
		registered := ""

		if instance.Registered {
			registered = " [Registered]"
		}

		fmt.Println(instance.Config.ServiceAddr.IPAddress + ":" + strconv.Itoa(instance.Config.ServiceAddr.Port) + " - " + instance.Config.Name + " " + instance.Config.Version + registered)
	}
}

func ListHosts(q *client.Query) {
	results := q.FindHosts()

	for _, host := range results {
		fmt.Println(host)
	}
}

func ListRegions(q *client.Query) {
	results := q.FindRegions()

	for _, region := range results {
		fmt.Println(region)
	}
}

func ListServices(q *client.Query) {
	results := q.FindServices()

	for _, service := range results {
		fmt.Println(service)
	}
}

func ListServiceVersions(q *client.Query) {
	if q.Service == "" {
		fmt.Println("Service name is required")
		return
	}

	results := q.FindServiceVersions()

	for _, version := range results {
		fmt.Println(version)
	}
}

func PrintTopology(q *client.Query) {
	topology := make(map[string]map[string]map[string]map[string][]*service.Service)

	results := q.FindInstances()

	// Build topology hash first
	for _, instance := range results {
		if topology[instance.Config.Region] == nil {
			topology[instance.Config.Region] = make(map[string]map[string]map[string][]*service.Service)
		}

		if topology[instance.Config.Region][instance.Config.ServiceAddr.IPAddress] == nil {
			topology[instance.Config.Region][instance.Config.ServiceAddr.IPAddress] = make(map[string]map[string][]*service.Service)
		}

		if topology[instance.Config.Region][instance.Config.ServiceAddr.IPAddress][instance.Config.Name] == nil {
			topology[instance.Config.Region][instance.Config.ServiceAddr.IPAddress][instance.Config.Name] = make(map[string][]*service.Service)
		}

		if topology[instance.Config.Region][instance.Config.ServiceAddr.IPAddress][instance.Config.Name][instance.Config.Version] == nil {
			topology[instance.Config.Region][instance.Config.ServiceAddr.IPAddress][instance.Config.Name][instance.Config.Version] = make([]*service.Service, 0)
		}

		topology[instance.Config.Region][instance.Config.ServiceAddr.IPAddress][instance.Config.Name][instance.Config.Version] = append(topology[instance.Config.Region][instance.Config.ServiceAddr.IPAddress][instance.Config.Name][instance.Config.Version], instance)
	}

	// Now we can print the correct heirarchy
	for regionName, region := range topology {
		fmt.Println("Region: " + regionName)

		for hostName, host := range region {
			fmt.Println("\tHost: " + hostName)

			for serviceName, service := range host {
				fmt.Println("\t\tService: " + serviceName)

				for versionName, version := range service {
					fmt.Println("\t\t\tVersion: " + versionName)

					for _, instance := range version {
						registered := ""

						if instance.Registered {
							registered = " [Registered]"
						}

						fmt.Println("\t\t\t\t" + instance.Config.ServiceAddr.IPAddress + ":" + strconv.Itoa(instance.Config.ServiceAddr.Port) + registered)
					}
				}
			}
		}
	}
}

func CommandLineHelp() {
	fmt.Print(`Usage: sky [options] command <arguments>

Commands:

	cli: Interactive shell for executing commands against skynet cluster
	hosts: List all hosts available that meet the specified criteria
		-service - limit results to hosts running the specified service
		-version - limit results to hosts running the specified version of the service (-service required)
		-region - limit results to hosts in the specified region
	instances: List all instances available that meet the specified criteria
		-service - limit results to instances of the specified service
		-version - limit results to instances of the specified version of service
		-region - limit results to instances in the specified region
		-host - limit results to instances on the specified host
		-port - limit results to instances on the specified port
		-registered - (true, false) limit results to instances that are registered (accepting requests)
	regions: List all regions available that meet the specified criteria
	services: List all services available that meet the specified criteria
		-host - limit results to the specified host
		-host - limit results to the specified port
		-region - limit results to hosts in the specified region
		-region - limit results to hosts in the specified region
	versions: List all services available that meet the specified criteria
		-service - service name (required)
		-host - limit results to the specified host
		-host - limit results to the specified port
		-region - limit results to hosts in the specified region
	topology: Print detailed heirarchy of regions/hosts/services/versions/instances
		-service - limit results to instances of the specified service
		-version - limit results to instances of the specified version of service
		-region - limit results to instances in the specified region
		-host - limit results to instances on the specified host
		-host - limit results to the specified port
	daemon [config file]: Run the "SkynetDaemon" service, and deploy services listed in the provided config
	remote [command]: Administer the services run by the SkynetDaemon service on the given host.
		-host - the SkynetDaemon service to connect to
		use "remote help" for details.

		
		

`)

}

/*
 * CLI Logic
 */

func InteractiveShell() {
	lineReader := bufio.NewReader(os.Stdin)
	doozer := Doozer()

	fmt.Println("Skynet Interactive Shell")
	prompt()

	query := &client.Query{
		DoozerConn: doozer,
	}

	for {
		l, _, e := lineReader.ReadLine()
		if e != nil {
			break
		}

		s := string(l)
		parts := strings.Split(s, " ")

		switch parts[0] {
		case "exit":
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

			fmt.Printf("Region: %v\nHost: %v\nService:%v\nVersion: %v\nRegistered: %v\n", query.Region, query.Host, query.Service, query.Version, registered)
		default:
			fmt.Println("Unknown Command - type 'help' for a list of commands")
		}

		prompt()
	}
}

func InteractiveShellHelp() {
	fmt.Print(`
Commands:
	hosts: List all hosts available that meet the specified criteria
	instances: List all instances available that meet the specified criteria
	regions: List all regions available that meet the specified criteria
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

func prompt() {
	fmt.Printf("> ")
}
