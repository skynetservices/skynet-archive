package main

import (
	"fmt"
)

type SubserviceDeployment struct {
	ServicePath string
	Args        string
}

func (sd SubserviceDeployment) String() string {
	return fmt.Sprintf("Deployed %s %s", sd.ServicePath, sd.Args)
}
