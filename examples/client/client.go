package main

import (
	"fmt"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/client"
	"os"
)

func main() {
	config, _ := skynet.GetClientConfigFromFlags(os.Args[1:]...)

	var err error
	config.Log = skynet.NewConsoleLogger("TestServiceClient", os.Stderr)

	client := client.NewClient(config)

	// This will not fail if no services currently exist, as connections are created on demand
	// this saves from chicken and egg issues with dependencies between services
	service := client.GetService("TestService", "", "", "") // any version, any region, any host

	// This on the other hand will fail if it can't find a service to connect to
	in := map[string]interface{}{
		"data": "Upcase me!!",
	}
	out := map[string]interface{}{}
	err = service.Send(nil, "Upcase", in, &out)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(out["data"].(string))

}
