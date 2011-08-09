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

func callRand(){
	service := "MyRandomService"
	client, _ := skylib.GetRandomClientByService(service)
	request := rand.Intn(10) + 1
	println("Request:", request)
	var response string = "default" // to prove it has changed
	client.Call(service+".RandString", request, &response)
	println("Reponse:", response)
	println("Done.")
}

func main() {

	// These two lines are the only lines required by an initiator
	flag.Parse()
	skylib.NewAgent().Start()

	for x:=1; x<100; x++{
		callRand()
	}

}
