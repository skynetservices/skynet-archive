package main

import (
	"github.com/bketelsen/skynet/skylib"
	"rand"
	"time"
	"flag"
	"fmt"
	"github.com/bketelsen/skynet/examples/GetWidgets/MyReqResp"
)

func init() {
	rand.Seed(time.Seconds())
}

func main() {
	flag.Parse()
	skylib.NewAgent().Start()
	sig := "UnisexService"
	println("Seeking services of type ", sig)
	client, err := skylib.GetRandomClientByService(sig)
	skylib.CheckError(&err)
	inputs := make(map[string]interface{})
	inputs["YourInputValue"] = "YourMappedInputValue"
	cr := MyReqResp.SkynetRequest{Params: inputs, Body: []byte("0")}
	fmt.Printf("Request:%v\n", cr)
	var response = MyReqResp.SkynetResponse{}
	err = client.Call(sig+".HandleMale", cr, &response)
	skylib.CheckError(&err)
	fmt.Printf("Reponse:%v\n", response)
	println("Done.")
}
