package main

import (
	"fmt"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/client"
	"github.com/bketelsen/skynet/examples/testing/sleeper"
	"time"
)

func main() {
	config, _ := skynet.GetClientConfig()
	config.MaxConnectionsToInstance = 5
	client := client.NewClient(config)

	service := client.GetService("Sleeper", "", "", "")

	service.SetTimeout(1*time.Second, 30*time.Second)

	req := sleeper.Request{
		Message:  "Hello!",
		Duration: 5 * time.Second,
	}
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
