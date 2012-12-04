package main

import (
	"flag"
	"fmt"
	"github.com/bketelsen/skynet"
	"os"
	"strconv"
)

var (
	flagset                 = flag.NewFlagSet("sky", flag.ExitOnError)
	VersionFlag     *string = flagset.String("version", "", "service version")
	ServiceNameFlag *string = flagset.String("service", "", "service name")
	PortFlag        *string = flagset.String("port", "", "port")
	RegisteredFlag  *string = flagset.String("registered", "", "registered")
)

var DC *skynet.DoozerConnection
var config skynet.ClientConfig

func main() {
	logger := skynet.NewConsoleSemanticLogger("Sky", os.Stdout)

	config = skynet.ClientConfig{
		DoozerConfig: &skynet.DoozerConfig{},
		Log:          logger,
	}

	skynet.FlagsForClient(&config, flagset)

	err := flagset.Parse(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		return
	}

	query := &skynet.Query{
		DoozerConn: Doozer(config.DoozerConfig),
		Service:    *ServiceNameFlag,
		Version:    *VersionFlag,
		Host:       config.Host,
		Region:     config.Region,
		Port:       *PortFlag,
	}

	fmt.Println(flagset.Args())

	switch flagset.Arg(0) {
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
	case "register":
		Register(query)
	case "unregister":
		Unregister(query)
	case "stop":
		Stop(query)
	case "restart":
		Restart(query)
	case "deploy":
		args := flagset.Args()
		fmt.Println(args)
		if len(args) < 2 {
			fmt.Println("Usage: deploy <service path> <args>")
			return
		}

		Deploy(query, args[1], args[2:]...)

	case "cli":
		InteractiveShell()

	default:
		CommandLineHelp()
	}
}

func Doozer(dcfg *skynet.DoozerConfig) *skynet.DoozerConnection {
	if DC == nil {
		DC = Connect(dcfg)
	}

	return DC
}

func Connect(dcfg *skynet.DoozerConfig) *skynet.DoozerConnection {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Failed to connect to Doozer")
			os.Exit(1)
		}
	}()

	// TODO: This needs to come from command line, or environment variable
	conn := skynet.NewDoozerConnection(dcfg.Uri, "", false, nil) // nil as the last param will default to a Stdout logger
	conn.Connect()

	return conn
}

func ListInstances(q *skynet.Query) {
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

func ListHosts(q *skynet.Query) {
	results := q.FindHosts()

	for _, host := range results {
		fmt.Println(host)
	}
}

func ListRegions(q *skynet.Query) {
	results := q.FindRegions()

	for _, region := range results {
		fmt.Println(region)
	}
}

func ListServices(q *skynet.Query) {
	results := q.FindServices()

	for _, service := range results {
		fmt.Println(service)
	}
}

func ListServiceVersions(q *skynet.Query) {
	if q.Service == "" {
		fmt.Println("Service name is required")
		return
	}

	results := q.FindServiceVersions()

	for _, version := range results {
		fmt.Println(version)
	}
}

func PrintTopology(q *skynet.Query) {
	topology := make(map[string]map[string]map[string]map[string][]*skynet.ServiceInfo)

	results := q.FindInstances()

	// Build topology hash first
	for _, instance := range results {
		if topology[instance.Config.Region] == nil {
			topology[instance.Config.Region] = make(map[string]map[string]map[string][]*skynet.ServiceInfo)
		}

		if topology[instance.Config.Region][instance.Config.ServiceAddr.IPAddress] == nil {
			topology[instance.Config.Region][instance.Config.ServiceAddr.IPAddress] = make(map[string]map[string][]*skynet.ServiceInfo)
		}

		if topology[instance.Config.Region][instance.Config.ServiceAddr.IPAddress][instance.Config.Name] == nil {
			topology[instance.Config.Region][instance.Config.ServiceAddr.IPAddress][instance.Config.Name] = make(map[string][]*skynet.ServiceInfo)
		}

		if topology[instance.Config.Region][instance.Config.ServiceAddr.IPAddress][instance.Config.Name][instance.Config.Version] == nil {
			topology[instance.Config.Region][instance.Config.ServiceAddr.IPAddress][instance.Config.Name][instance.Config.Version] = make([]*skynet.ServiceInfo, 0)
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
		-port - limit results to the specified port
		-region - limit results to hosts in the specified region
		-region - limit results to hosts in the specified region
	versions: List all services available that meet the specified criteria
		-service - service name (required)
		-host - limit results to the specified host
		-port - limit results to the specified port
		-region - limit results to hosts in the specified region
	topology: Print detailed heirarchy of regions/hosts/services/versions/instances
		-service - limit results to instances of the specified service
		-version - limit results to instances of the specified version of service
		-region - limit results to instances in the specified region
		-host - limit results to instances on the specified host
		-port - limit results to the specified port
	deploy: deploy new instances to cluster (deploy <service path> <args>)
		-region - deploy only to the specified region
		-host - deploy to the specified host
	stop: Stop all instances available that meet the specified criteria
		-service - limit command to instances of the specified service
		-version - limit command to instances of the specified version of service
		-region - limit command to instances in the specified region
		-host - limit command to instances on the specified host
		-port - limit command to instances on the specified port
		-registered - (true, false) limit command to instances that are registered (accepting requests)
	register: Register all instances available that meet the specified criteria
		-service - limit command to instances of the specified service
		-version - limit command to instances of the specified version of service
		-region - limit command to instances in the specified region
		-host - limit command to instances on the specified host
		-port - limit command to instances on the specified port
		-registered - (true, false) limit command to instances that are registered (accepting requests)
	unregister: Unregister all instances available that meet the specified criteria
		-service - limit command to instances of the specified service
		-version - limit command to instances of the specified version of service
		-region - limit command to instances in the specified region
		-host - limit command to instances on the specified host
		-port - limit command to instances on the specified port
		-registered - (true, false) limit command to instances that are registered (accepting requests)
		
		

`)

}
