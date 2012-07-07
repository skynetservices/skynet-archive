package main

import (
	"fmt"
	"github.com/bketelsen/skynet/skylib"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	c := make(chan os.Signal, 1)

	config := &skylib.ClientConfig{
		DoozerConfig: &skylib.DoozerConfig{
			Uri:          "127.0.0.1:8046",
			AutoDiscover: true,
		},
	}

	var err error
	config.Log, err = skylib.NewMongoLogger("localhost", "skynet", "log")
	if err != nil {
		config.Log = skylib.NewConsoleLogger(os.Stderr)
		config.Log.Item("Could not connect to mongo db for logging")
	}

	client := skylib.NewClient(config)

	// This will not fail if no services currently exist, as connections are created on demand
	// this saves from chicken and egg issues with dependencies between services
	service := client.GetService("TestService", "", "", "") // any version, any region, any host

	// This on the other hand will fail if it can't find a service to connect to
	ret, err := service.Send("Upcase", "Upcase me!!")

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(ret)

	watchSignals(c)
}

func watchSignals(c chan os.Signal) {
	signal.Notify(c, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGSEGV, syscall.SIGSTOP, syscall.SIGTERM)

	for {
		select {
		case sig := <-c:
			switch sig.(syscall.Signal) {
			// Trap signals for clean shutdown
			case syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGSEGV, syscall.SIGSTOP, syscall.SIGTERM:
				syscall.Exit(0)
			}
		}
	}
}
