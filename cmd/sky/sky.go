package main

import (
	"flag"
	"fmt"
	"github.com/skynetservices/skynet2"
	"github.com/skynetservices/zkmanager"
	"os"
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
		ListHosts(args)
	}
}

func ListHosts(args []string) {
	flagset := flag.NewFlagSet("deploy", flag.ExitOnError)
	services := flagset.String("services", "", "services")
	regions := flagset.String("regions", "", "regions")
	registered := flagset.String("registered", "", "registered")

	flagsetArgs, _ := skynet.SplitFlagsetFromArgs(flagset, args)

	err := flagset.Parse(flagsetArgs)
	if err != nil {
		panic(err)
		return
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

	hosts, err := skynet.GetServiceManager().ListHosts(skynet.Criteria{
		Regions:    regionCriteria,
		Registered: reg,
		Services:   serviceCriteriaFromCsv(*services),
	})

	if err != nil {
		panic(err)
	}

	for _, h := range hosts {
		fmt.Println(h)
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
            build: Uses build.cfg or optional config to build the current project
                  -config - config file to use
            deploy: Uses build.cfg or optional config to deploy the current project
                  -config - config file to use
            hosts: List all hosts available that meet the specified criteria
            -services - limit results to hosts running the specified comma separated services example. -services=MyService or --services=MyService:v1
                  -region - limit results to hosts in the specified comma separated regions
                  -registered - limit results to hosts that have registered instances
            
            
  `)
}
