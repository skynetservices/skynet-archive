package main

import (
	"github.com/kballard/go-shellquote"
	"os/exec"
)

type Terminal interface {
	Exec(cmd string) (out []byte, err error)
}

type LocalTerminal struct{}

func (t *LocalTerminal) Exec(cmd string) (out []byte, err error) {
	args, err := shellquote.Split(cmd)

	if err != nil {
		return
	}

	command := exec.Command(args[0], args[1:]...)

	return command.CombinedOutput()
}
