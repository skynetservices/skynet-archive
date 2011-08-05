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
	return
}

func main() {
	flag.Parse()
	sig := &MyRandomService{}
	skylib.NewAgent().Register(sig).Start().Wait()
}
