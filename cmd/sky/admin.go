package main

import (
	"fmt"
	"github.com/kballard/go-shellquote"
	"github.com/skynetservices/skynet2"
	"github.com/skynetservices/skynet2/daemon"
	"os"
	"text/template"
)

var deployTemplate = template.Must(template.New("").Parse(
	`Deployed service with UUID {{.UUID}}.
`))

func Start(criteria *skynet.Criteria, args []string) {
	if len(args) < 1 {
		fmt.Println("Please provide a service name 'sky start binaryName'")
		return
	}

	hosts, err := skynet.GetServiceManager().ListHosts(criteria)

	if err != nil {
		panic(err)
	}

	for _, host := range hosts {
		fmt.Println("deploying to host: " + host)
		d := daemon.GetDaemonForHost(Client, host)

		in := daemon.StartRequest{
			BinaryName: args[0],
			Args:       shellquote.Join(args[1:]...),
		}
		out, err := d.Start(in)

		if err != nil {
			fmt.Println("Returned Error: " + err.Error())
			return
		}

		deployTemplate.Execute(os.Stdout, out)
	}
}

func Stop(criteria *skynet.Criteria) {
}

func Restart(criteria *skynet.Criteria) {
}

func Register(criteria *skynet.Criteria) {
}

func Unregister(criteria *skynet.Criteria) {
}
