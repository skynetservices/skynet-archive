package main

import (
	"github.com/bketelsen/skynet/skylib"
	"os"
	"flag"
)

type MyRandomProvision struct {

}

func (*MyRandomProvision) RandString(n int, response *string) (err os.Error) {
	word := skylib.RandWord(n)
	skylib.LogError("RandString:", n, word)
	*response = word
	skylib.LogError(*response)
	return
}

func main() {
	flag.Parse()
	node := skylib.NewNode()
	r := &MyRandomProvision{}
	node.RegisterRpcServer(r).Start().Wait()
}
