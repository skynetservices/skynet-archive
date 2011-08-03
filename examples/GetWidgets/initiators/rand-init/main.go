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
	skylib.NewNode().Start()
	prov := "MyRandomProvision"
	println("Seeking services of type ", prov)
	client, err := skylib.GetRandomClientByProvides(prov)
	skylib.CheckError(&err)
	request := rand.Intn(10) + 1
	println("Request:", request)
	var response string = "default" // to prove it has changed
	err = client.Call(prov+".RandString", request, &response)
	skylib.CheckError(&err)
	println("Reponse:", response)
	println("Done.")
}
