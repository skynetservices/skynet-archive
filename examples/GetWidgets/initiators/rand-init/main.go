package main

import (
	"github.com/bketelsen/skynet/skylib"
	"rand"
	"time"
	"flag"
	"strconv"
)

func init() {
	rand.Seed(time.Seconds())
}

func callRand() {
	service := "MyRandomService"
	client, _ := skylib.GetRandomClientByService(service)
	request := rand.Intn(10) + 1
	println("Request:", request)
	var response string = "default" // to prove it has changed
	client.Call(service+".RandString", request, &response)
	println("Reponse:", response)
	println("Done.")
}

func run(n uint) {
	// Ordinarily, we would use the same Service everytime, but
	// here we are just testing.
	for ; n > 0; n-- {
		callRand()
	}
}

func main() {

	// These two lines are the only lines required by an initiator
	flag.Parse()
	skylib.NewAgent().Start()

	var n uint = 1
	if flag.NArg() > 0 {
		n, _ = strconv.Atoui(flag.Arg(0))
	}
	run(n)
}
