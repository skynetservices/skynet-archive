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
	sig := "MyRandomService"
	println("Seeking services of type ", sig)
	client, err := skylib.GetRandomClientByService(sig)
	skylib.CheckError(&err)
	request := rand.Intn(10) + 1
	println("Request:", request)
	var response string = "default" // to prove it has changed
	err = client.Call(sig+".RandString", request, &response)
	skylib.CheckError(&err)
	println("Reponse:", response)
	println("Done.")
}
