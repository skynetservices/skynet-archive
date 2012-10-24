package main

import (
	"flag"
	"fmt"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/client"
	"github.com/bketelsen/skynet/examples/testing/sleeper"
	"os"
	"time"
)

func main() {
	config := &skynet.ClientConfig{
		DoozerConfig: &skynet.DoozerConfig{},
	}

	flagset := flag.NewFlagSet("sleepclient", flag.ContinueOnError)

	skynet.FlagsForClient(config, flagset)

	req := sleeper.Request{
		Message: "Hello!",
	}

	var (
		retry  time.Duration
		giveup time.Duration
	)

	flagset.DurationVar(&req.Duration, "sleepfor", 5*time.Second,
			"how long to sleep")
	flagset.BoolVar(&req.ExitWhenDone, "exit", false,
			"have the service call os.Exit(0) when finished sleeping")
	flagset.BoolVar(&req.PanicWhenDone, "panic", false,
			"have the service panic when finished sleeping")
	flagset.BoolVar(&req.UnregisterWhenDone, "unregister", false,
			"have the service unregister when finished sleeping")
	flagset.BoolVar(&req.UnregisterHalfwayThrough, "unregister-halfway", false,
			"have the service unregister half-way through the sleep")
	flagset.DurationVar(&retry, "retry", time.Second,
			"how long to wait before trying again")
	flagset.DurationVar(&giveup, "giveup", 5*time.Second,
			"how long to wait before giving up")

	flagset.Parse(os.Args[1:])

	config.MaxConnectionsToInstance = 5

	client := client.NewClient(config)

	service := client.GetService("Sleeper", "", "", "")

	service.SetTimeout(retry, giveup)

	resp := sleeper.Response{}

	start := time.Now()

	err := service.Send(nil, "Sleep", req, &resp)

	duration := time.Now().Sub(start).Nanoseconds()

	fmt.Printf("request took %dns\n", duration)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%s -> %s\n", req.Message, resp.Message)
	}
}
