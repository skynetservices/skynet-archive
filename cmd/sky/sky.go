package main

import (
	"flag"
	"fmt"
	"github.com/skynetservices/skynet2"
	"github.com/skynetservices/zkmanager"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	skynet.SetServiceManager(zkmanager.NewZookeeperServiceManager(os.Getenv("SKYNET_ZOOKEEPER"), 1*time.Second))

	args := os.Args[1:]

	if len(args) == 0 {
		CommandLineHelp()
		return
	}

	switch args[0] {
	case "help", "h":
		CommandLineHelp()
	case "build", "b":
		flagset := flag.NewFlagSet("build", flag.ExitOnError)
		config := flagset.String("config", "./build.cfg", "build config file")
		flagsetArgs, _ := skynet.SplitFlagsetFromArgs(flagset, args)

		err := flagset.Parse(flagsetArgs)
		if err != nil {
			panic(err)
			return
		}

		Build(*config)
	case "deploy", "d":
		flagset := flag.NewFlagSet("deploy", flag.ExitOnError)
		config := flagset.String("config", "./build.cfg", "build config file")
		flagsetArgs, _ := skynet.SplitFlagsetFromArgs(flagset, args)

		err := flagset.Parse(flagsetArgs)
		if err != nil {
			panic(err)
			return
		}

		Deploy(*config)
	case "hosts":
		ListHosts(criteriaFromArgs(args))
	case "regions":
		ListRegions(criteriaFromArgs(args))
	case "services":
		ListServices(criteriaFromArgs(args))
	case "versions":
		ListVersions(criteriaFromArgs(args))
	case "instances":
		ListInstances(criteriaFromArgs(args))
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
		panic(err)
	}

	return regions
}

func ListVersions(c *skynet.Criteria) {
	printList(getVersions(c))
}

func getVersions(c *skynet.Criteria) []string {
	versions, err := skynet.GetServiceManager().ListVersions(c)

	if err != nil {
		panic(err)
	}

	return versions
}

func ListServices(c *skynet.Criteria) {
	printList(getServices(c))
}

func getServices(c *skynet.Criteria) []string {
	services, err := skynet.GetServiceManager().ListServices(c)

	if err != nil {
		panic(err)
	}

	return services
}

func ListHosts(c *skynet.Criteria) {
	printList(getHosts(c))
}

func getHosts(c *skynet.Criteria) []string {
	hosts, err := skynet.GetServiceManager().ListHosts(c)

	if err != nil {
		panic(err)
	}

	return hosts
}

func ListInstances(c *skynet.Criteria) {
	for _, instance := range getInstances(c) {
		fmt.Println(instance.ServiceAddr.String() + " - " + instance.Name + " " + instance.Version + " " + strconv.FormatBool(instance.Registered))
	}
}

func getInstances(c *skynet.Criteria) []skynet.ServiceInfo {
	instances, err := skynet.GetServiceManager().ListInstances(c)

	if err != nil {
		panic(err)
	}

	return instances
}

func printList(list []string) {
	for _, v := range list {
		fmt.Println(v)
	}
}

func criteriaFromArgs(args []string) *skynet.Criteria {
	flagset := flag.NewFlagSet("deploy", flag.ExitOnError)
	services := flagset.String("services", "", "services")
	regions := flagset.String("regions", "", "regions")
	registered := flagset.String("registered", "", "registered")

	flagsetArgs, _ := skynet.SplitFlagsetFromArgs(flagset, args)

	err := flagset.Parse(flagsetArgs)
	if err != nil {
		panic(err)
	}

	regionCriteria := make([]string, 0, 0)

	if len(*regions) > 0 {
		regionCriteria = strings.Split(*regions, ",")
	}

	var reg *bool
	if *registered == "true" {
		*reg = true
	} else if *registered == "false" {
		*reg = false
	}

	return &skynet.Criteria{
		Regions:    regionCriteria,
		Registered: reg,
		Services:   serviceCriteriaFromCsv(*services),
	}
}

func serviceCriteriaFromCsv(csv string) (criteria []skynet.ServiceCriteria) {
	if csv == "" {
		return
	}

	services := strings.Split(csv, ",")

	for _, s := range services {
		parts := strings.Split(s, ":")

		c := skynet.ServiceCriteria{
			Name: parts[0],
		}

		if len(parts) > 1 {
			c.Version = parts[1]
		}

		criteria = append(criteria, c)
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
            regions: List all regions available that meet the specified criteria
                  -services - limit results to regions running the specified comma separated services example. -services=MyService or --services=MyService:v1
                  -hosts - limit results to regions with the specified comma separated hosts
                  -registered - limit results to regions that have registered instances
            hosts: List all hosts available that meet the specified criteria
                  -services - limit results to hosts running the specified comma separated services example. -services=MyService or --services=MyService:v1
                  -regions - limit results to hosts in the specified comma separated regions
                  -registered - limit results to hosts that have registered instances
            services: List all services available that meet the specified criteria
                  -hosts - limit results to services running on the specified comma separated hosts
                  -regions - limit results to services in the specified comma separated regions
                  -registered - limit results to services that have registered instances
            versions: List all versions available that meet the specified criteria
                  -hosts - limit results to versions running on the specified comma separated hosts
                  -regions - limit results to versions in the specified comma separated regions
                  -registered - limit results to versions that have registered instances
                  -services - limit results to versions running the specified comma separated services example. -services=MyService or --services=MyService:v1
            instances: List all instances available that meet the specified criteria
                  -hosts - limit results to instances running on the specified comma separated hosts
                  -regions - limit results to instances in the specified comma separated regions
                  -registered - limit results to instances that have registered instances
                  -services - limit results to instances running the specified comma separated services example. -services=MyService or --services=MyService:v1
            
            
  `)
}
