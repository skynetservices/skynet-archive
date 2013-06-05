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
		Build()
	}
}

func CommandLineHelp() {
	fmt.Print(`Usage: sky [options] command <arguments>

    Commands:
            build: Uses build.cfg to build the current project
            
            
            
  `)
}
