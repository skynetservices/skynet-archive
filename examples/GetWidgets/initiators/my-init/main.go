package main

import (
	"github.com/bketelsen/skynet/skylib"
	"rand"
	"time"
	"flag"
	"fmt"
	"syscall"
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
	for n:=0;; n++ {
		println(n)
		for i:=0; i<3; i++ {
			is := fmt.Sprintf("%d", i)
			cr := MyReqResp.SkynetRequest{Params: inputs, Body: []byte(is)}
			fmt.Printf("Request:%v\n", cr)
			var response = MyReqResp.SkynetResponse{}
			err = client.Call(sig+".HandleMale", cr, &response)
			println("MaleResponse:", string(response.Body))
			err = client.Call(sig+".HandleFemale", cr, &response)
			println("FemaleResponse:", string(response.Body))
			skylib.CheckError(&err)
		}
		syscall.Sleep(1e9)
	}
	println("Done.")
}
