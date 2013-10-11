package main

import (
	"flag"
	"fmt"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/config"
	"github.com/skynetservices/skynet/log"
	_ "github.com/skynetservices/zkmanager"
	"os"
	"strconv"
	"strings"
)

func main() {
	log.SetLogLevel(log.ERROR)

	var args []string
	criteria, args := criteriaFromArgs(os.Args[1:])

	if len(args) == 0 {
		CommandLineHelp()
		return
	}

	switch args[0] {
	case "help", "h":
		CommandLineHelp()
	case "build", "b":
		flagset := flag.NewFlagSet("build", flag.ExitOnError)
		configFile := flagset.String("build", "./build.cfg", "build config file")
		flagsetArgs, _ := config.SplitFlagsetFromArgs(flagset, args)

		err := flagset.Parse(flagsetArgs)
		if err != nil {
			log.Fatal(err)
			return
		}

		Build(*configFile)
	case "deploy", "d":
		flagset := flag.NewFlagSet("deploy", flag.ExitOnError)
		configFile := flagset.String("build", "./build.cfg", "build config file")
		flagsetArgs, _ := config.SplitFlagsetFromArgs(flagset, args)

		err := flagset.Parse(flagsetArgs)
		if err != nil {
			log.Fatal(err)
			return
		}

		Deploy(*configFile, criteria)
	case "hosts":
		ListHosts(criteria)
	case "regions":
		ListRegions(criteria)
	case "services":
		ListServices(criteria)
	case "versions":
		ListVersions(criteria)
	case "instances":
		ListInstances(criteria)
	case "start":
		Start(criteria, args[1:])
	case "stop":
		Stop(criteria)
	case "restart":
		Restart(criteria)
	case "register":
		Register(criteria)
	case "unregister":
		Unregister(criteria)
	case "log":
		SetLogLevel(criteria, args[1])
	case "daemon":
		if len(args) >= 2 {
			switch args[1] {
			case "log":
				if len(args) >= 3 {
					SetDaemonLogLevel(criteria, args[2])
				} else {
					fmt.Println("Must supply a log level")
				}
			case "stop":
				StopDaemon(criteria)
			}
		} else {
			fmt.Println("Supported subcommands for daemon are log, and stop")
		}
	case "cli":
		InteractiveShell()
	default:
		fmt.Println("Unknown Command: ", args[0])
		CommandLineHelp()
	}
}

func ListRegions(c *skynet.Criteria) {
	printList(getRegions(c))
}

func getRegions(c *skynet.Criteria) []string {
	regions, err := skynet.GetServiceManager().ListRegions(c)

	if err != nil {
		log.Fatal(err)
	}

	return regions
}

func ListVersions(c *skynet.Criteria) {
	printList(getVersions(c))
}

func getVersions(c *skynet.Criteria) []string {
	versions, err := skynet.GetServiceManager().ListVersions(c)

	if err != nil {
		log.Fatal(err)
	}

	return versions
}

func ListServices(c *skynet.Criteria) {
	printList(getServices(c))
}

func getServices(c *skynet.Criteria) []string {
	services, err := skynet.GetServiceManager().ListServices(c)

	if err != nil {
		log.Fatal(err)
	}

	return services
}

func ListHosts(c *skynet.Criteria) {
	printList(getHosts(c))
}

func getHosts(c *skynet.Criteria) []string {
	hosts, err := skynet.GetServiceManager().ListHosts(c)

	if err != nil {
		log.Fatal(err)
	}

	return hosts
}

func ListInstances(c *skynet.Criteria) {
	for _, instance := range getInstances(c) {
		fmt.Println(instance.ServiceAddr.String() + " - " + instance.Region + " - " + instance.Name + " " + instance.Version + " " + strconv.FormatBool(instance.Registered) + " (" + instance.UUID + ")")
	}
}

func getInstances(c *skynet.Criteria) []skynet.ServiceInfo {
	instances, err := skynet.GetServiceManager().ListInstances(c)

	if err != nil {
		log.Fatal(err)
	}

	return instances
}

func printList(list []string) {
	for _, v := range list {
		fmt.Println(v)
	}
}

func criteriaFromArgs(args []string) (*skynet.Criteria, []string) {
	flagset := flag.NewFlagSet("deploy", flag.ExitOnError)
	services := flagset.String("services", "", "services")
	regions := flagset.String("regions", "", "regions")
	instances := flagset.String("instances", "", "instances")
	hosts := flagset.String("hosts", "", "hosts")
	registered := flagset.String("registered", "", "registered")

	flagsetArgs, args := config.SplitFlagsetFromArgs(flagset, args)

	err := flagset.Parse(flagsetArgs)
	if err != nil {
		log.Fatal(err)
	}

	regionCriteria := make([]string, 0, 0)

	if len(*regions) > 0 {
		regionCriteria = strings.Split(*regions, ",")
	}

	hostCriteria := make([]string, 0, 0)

	if len(*hosts) > 0 {
		hostCriteria = strings.Split(*hosts, ",")
	}

	instanceCriteria := make([]string, 0, 0)

	if len(*instances) > 0 {
		instanceCriteria = strings.Split(*instances, ",")
	}

	var reg *bool
	if *registered == "true" {
		*reg = true
	} else if *registered == "false" {
		*reg = false
	}

	return &skynet.Criteria{
		Regions:    regionCriteria,
		Hosts:      hostCriteria,
		Instances:  instanceCriteria,
		Registered: reg,
		Services:   serviceCriteriaFromCsv(*services),
	}, args
}

func serviceCriteriaFromCsv(csv string) (criteria []skynet.ServiceCriteria) {
	if csv == "" {
		return
	}

	services := strings.Split(csv, ",")

	for _, s := range services {
		criteria = append(criteria, serviceCriteriaFromString(s))
	}

	return
}

func serviceCriteriaFromString(s string) (c skynet.ServiceCriteria) {
	parts := strings.Split(s, ":")

	c = skynet.ServiceCriteria{
		Name: parts[0],
	}

	if len(parts) > 1 {
		c.Version = parts[1]
	}

	return
}

func CommandLineHelp() {
	fmt.Print(`Usage: sky [options] command <arguments>

    Commands:
            cli: Interactive shell for executing commands against skynet cluster
            build: Uses build.cfg or optional config to build the current project
                  -config - config file to use
            deploy: Uses build.cfg or optional config to deploy the current project
                  -config - config file to use

            log: Set change log level of service that meet the specified criteria log <level>, options are DEBUG, TRACE, INFO, WARN, FATAL, PANIC
                  -hosts - change log level for services only on the specified comma separated hosts
                  -regions - change log level for services in the specified comma separated regions
                  -registered - change log level only on hosts that have registered instances
                  -services - change log level only on hosts that are running the specified comma separated services example. -services=MyService or --services=MyService:v1
                  -instances - change log level only on hosts that have the specified instances on them

            daemon log: Set change log level of daemons that meet the specified criteria daemon log <level>, options are DEBUG, TRACE, INFO, WARN, FATAL, PANIC
                  -hosts - change log level for daemons only on the specified comma separated hosts
                  -regions - change log level for daemons in the specified comma separated regions
                  -registered - change log level only on daemons that have registered instances
                  -services - change log level only on daemons that are running the specified comma separated services example. -services=MyService or --services=MyService:v1
                  -instances - change log level only on daemons that have the specified instances on them
            daemon stop: Stop daemons that meet the specified criteria daemon log <level>, options are DEBUG, TRACE, INFO, WARN, FATAL, PANIC
                  -hosts - stop daemons only on the specified comma separated hosts
                  -regions - stop daemons in the specified comma separated regions
                  -registered - stop daemons that have registered instances
                  -services - stop daemons that are running the specified comma separated services example. -services=MyService or --services=MyService:v1
                  -instances - stop daemons that have the specified instances on them

            start: Start named service on all hosts that match the supplied criteria "start <flags> <binaryName>"
                  -hosts - start service only on the specified comma separated hosts
                  -regions - start service only in the specified comma separated regions
                  -registered - start service only on hosts that have registered instances
                  -services - start service only on hosts that are running the specified comma separated services example. -services=MyService or --services=MyService:v1
                  -instances - start service on hosts that have the specified instances on them
            stop: Stop services that match the supplied criteria
                  -hosts - start service only on the specified comma separated hosts
                  -regions - start service only in the specified comma separated regions
                  -registered - start service only on hosts that have registered instances
                  -services - start service only on hosts that are running the specified comma separated services example. -services=MyService or --services=MyService:v1
                  -instances - start service on hosts that have the specified instances on them
            restart: Restart services that match the supplied criteria
                  -hosts - start service only on the specified comma separated hosts
                  -regions - start service only in the specified comma separated regions
                  -registered - start service only on hosts that have registered instances
                  -services - start service only on hosts that are running the specified comma separated services example. -services=MyService or --services=MyService:v1
                  -instances - start service on hosts that have the specified instances on them
            register: Register services that match the supplied criteria
                  -hosts - start service only on the specified comma separated hosts
                  -regions - start service only in the specified comma separated regions
                  -registered - start service only on hosts that have registered instances
                  -services - start service only on hosts that are running the specified comma separated services example. -services=MyService or --services=MyService:v1
                  -instances - start service on hosts that have the specified instances on them
            unregister: Unregister services that match the supplied criteria
                  -hosts - start service only on the specified comma separated hosts
                  -regions - start service only in the specified comma separated regions
                  -registered - start service only on hosts that have registered instances
                  -services - start service only on hosts that are running the specified comma separated services example. -services=MyService or --services=MyService:v1
                  -instances - start service on hosts that have the specified instances on them

            regions: List all regions available that meet the specified criteria
                  -services - limit results to regions running the specified comma separated services example. -services=MyService or --services=MyService:v1
                  -hosts - limit results to regions with the specified comma separated hosts
                  -registered - limit results to regions that have registered instances
                  -instances - limit results to regions that have instances specified
            hosts: List all hosts available that meet the specified criteria
                  -services - limit results to hosts running the specified comma separated services example. -services=MyService or --services=MyService:v1
                  -regions - limit results to hosts in the specified comma separated regions
                  -registered - limit results to hosts that have registered instances
                  -instances - limit results to hosts that have instances specified
            services: List all services available that meet the specified criteria
                  -hosts - limit results to services running on the specified comma separated hosts
                  -regions - limit results to services in the specified comma separated regions
                  -registered - limit results to services that have registered instances
                  -instances - limit results to services that have instances specified
            versions: List all versions available that meet the specified criteria
                  -hosts - limit results to versions running on the specified comma separated hosts
                  -regions - limit results to versions in the specified comma separated regions
                  -registered - limit results to versions that have registered instances
                  -services - limit results to versions running the specified comma separated services example. -services=MyService or --services=MyService:v1
                  -instances - limit results to versions that have instances specified
            instances: List all instances available that meet the specified criteria
                  -hosts - limit results to instances running on the specified comma separated hosts
                  -regions - limit results to instances in the specified comma separated regions
                  -registered - limit results to instances that have registered instances
                  -services - limit results to instances running the specified comma separated services example. -services=MyService or --services=MyService:v1
            
            
  `)
}
