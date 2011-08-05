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
	skylib.LogError("RandString:", n, word)
	*response = word
	skylib.LogError(*response)
	return
}

func main() {
	flag.Parse()
	agent := skylib.NewAgent()
	sig := &MyRandomService{}
	agent.Register(sig).Start().Wait()
}
