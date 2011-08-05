package main

import (
	"github.com/bketelsen/skynet/skylib"
	"rand"
	"time"
	"flag"
)

func init() {
	rand.Seed(time.Seconds())
}

func main() {
	flag.Parse()
	skylib.NewAgent().Start()
	service := "MyRandomService"
	client, _ := skylib.GetRandomClientByService(service)
	request := rand.Intn(10) + 1
	println("Request:", request)
	var response string = "default" // to prove it has changed
	client.Call(service+".RandString", request, &response)
	println("Reponse:", response)
	println("Done.")
}
