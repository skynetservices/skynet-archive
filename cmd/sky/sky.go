package main

import (
	"fmt"
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
		var config string

		if len(args) >= 2 {
			config = args[1]
		}

		Build(config)
	case "deploy", "d":
		var config string

		if len(args) >= 2 {
			config = args[1]
		}

		Deploy(config)
	}
}

func CommandLineHelp() {
	fmt.Print(`Usage: sky [options] command <arguments>

    Commands:
            build: Uses build.cfg or optional config to build the current project
            deploy: Uses build.cfg or optional config to deploy the current project
            
            
            
  `)
}
