package main

import (
	"github.com/bketelsen/skynet/skylib"
	"os"
	"flag"
)

type MyRandomService struct {

}

func (*MyRandomService) RandString(n int, response *string) (err os.Error) {
	word := skylib.RandWord(n)
	skylib.LogInfo("RandString:", n, word)
	*response = word
	skylib.LogInfo(*response)
	skylib.Requests.Add(1)
	return
}

func main() {
	flag.Parse()
	agent := skylib.NewAgent()
	sig := &MyRandomService{}
	agent.Register(sig).Start().Wait()
}
