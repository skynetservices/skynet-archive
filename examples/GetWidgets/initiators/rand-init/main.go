package main
import (
	"github.com/bketelsen/skynet/skylib"
	"rand"
	"syscall"
)

func main() {
	var tv syscall.Timeval
	syscall.Gettimeofday(&tv)
	//rand.Seed(tv.Nanoseconds())  // Does not compile! Why not?
	rand.Seed(int64(tv.Usec + 9999))
	skylib.Setup("Initiator.Rand")  // Only to get default Config.
	prov := "MyService.RandString"
	skylib.LogError("WhereAmI")
	skylib.LogError(prov)
	client, err := skylib.GetRandomClientByProvides(prov)
	skylib.LogError("hereAmI")
	skylib.CheckError(&err)
	request := rand.Intn(10) + 1
	println("Request:", request)
	var response string = "default"  // to prove it has changed
	err = client.Call("MyService.RandString", request, &response)
	skylib.CheckError(&err)
	println("Reponse:", response)
	println("Done.")
}
