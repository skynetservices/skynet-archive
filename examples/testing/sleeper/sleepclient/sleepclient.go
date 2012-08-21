package main

import (
	"fmt"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/client"
	"github.com/bketelsen/skynet/examples/testing/sleeper"
)

func main() {
	config, _ := skynet.GetClientConfigFromFlags()
	client := client.NewClient(config)

	service := client.GetService("Sleeper", "", "", "")

	req := sleeper.Request{
		Message: "Hello!",
	}
	resp := sleeper.Response{}
	err := service.Send(nil, "Sleep", req, &resp)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%s -> %s\n", req.Message, resp.Message)
	}
}
