package main

import (
	"fmt"
)

type SubserviceStart struct {
	BinaryName string
	Args       string
}

func (sd SubserviceStart) String() string {
	return fmt.Sprintf("Started %s %s", sd.BinaryName, sd.Args)
}
