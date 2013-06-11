package main

import (
	"flag"
	"fmt"
	"github.com/skynetservices/skynet2"
	"os"
)

func main() {
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
		config := flagset.String("version", "", "service version")
		flagsetArgs, _ := skynet.SplitFlagsetFromArgs(flagset, args)

		err := flagset.Parse(flagsetArgs)
		if err != nil {
			panic(err)
			return
		}

		Build(*config)
	case "deploy", "d":
		flagset := flag.NewFlagSet("deploy", flag.ExitOnError)
		config := flagset.String("version", "", "service version")
		flagsetArgs, _ := skynet.SplitFlagsetFromArgs(flagset, args)

		err := flagset.Parse(flagsetArgs)
		if err != nil {
			panic(err)
			return
		}

		Deploy(*config)
	}
}

func CommandLineHelp() {
	fmt.Print(`Usage: sky [options] command <arguments>

    Commands:
            build: Uses build.cfg or optional config to build the current project
                  -config - config file to use
            deploy: Uses build.cfg or optional config to deploy the current project
                  -config - config file to use
            
            
  `)
}
