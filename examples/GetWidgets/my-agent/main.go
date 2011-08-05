package main

import (
	"github.com/bketelsen/skynet/skylib"
	"github.com/bketelsen/skynet/examples/GetWidgets/MyReqResp"
	"flag"
)


func main() {
	flag.Parse()

	agent := skylib.NewAgent()
	sig := &MyReqResp.UnisexService{}
	agent.Register(sig)
	agent.Start().Wait()
}

