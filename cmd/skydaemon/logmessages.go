package main

import (
	"fmt"
)

type SubserviceStart struct {
	ServicePath string
	Args        string
}

func (sd SubserviceStart) String() string {
	return fmt.Sprintf("Started %s %s", sd.ServicePath, sd.Args)
}
