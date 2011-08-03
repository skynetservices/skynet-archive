package main
import (
	"github.com/bketelsen/skynet/skylib"
	"os"
	"flag"
//	"log"
	//"rpc"
)

type MyService struct {
}

func (*MyService) RandString(n int, response *string) (err os.Error) {
	word := skylib.RandWord(n)
	skylib.LogError("RandString:", n, word)
	*response = word
	skylib.LogError(*response)
	return
}

func main() {
	flag.Parse()
	skylib.Setup("MyService.RandString")
	r := &MyService{}
	//rpc.Register(r)
	server := skylib.NewRpcServer(r)
	server.Serve()
}
