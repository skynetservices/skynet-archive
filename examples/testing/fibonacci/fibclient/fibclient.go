package main

import (
	"fmt"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/client"
	"github.com/bketelsen/skynet/examples/testing/fibonacci"
	"os"
	"strconv"
)

func main() {
	config, args := skynet.GetClientConfigFromFlags(os.Args...)
	client := client.NewClient(config)

	service := client.GetService("Fibonacci", "", "", "")

	for _, arg := range args[1:] {
		index, err := strconv.Atoi(arg)
		if err != nil {
			panic(err)
		}
		req := fibonacci.Request{
			Index: index,
		}
		resp := fibonacci.Response{}
		err = service.Send(nil, "Index", req, &resp)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("%d -> %d\n", index, resp.Value)
		}
	}
}
