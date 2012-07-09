package main

import (
	"flag"
	"github.com/bketelsen/skynet/skylib"
	"fmt"
)

// Deploy() will run and maintain skynet services.
//
// Deploy() will initially deploy those specified in the file given in the "-config" option
//
// Deploy() will run the "SkynetDeployment" service, which can be used to remotely spawn
// new services on the host.
func Deploy(q *skylib.Query) {
	fmt.Println(flag.Args())
}
