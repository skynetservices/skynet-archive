package main

import (
	"github.com/bketelsen/skynet/skylib"
  "fmt"
)

func main() {
  config := &skylib.ClientConfig {
    DoozerConfig: &skylib.DoozerConfig {
      Uri: "127.0.0.1:8046",
    },
  }

	client := skylib.NewClient(config)
  service := client.GetService("TestService", "", "", "") // any version, any region, any host

  ret, _ := service.Send("Upcase", "foo")

  fmt.Println(ret)
}
