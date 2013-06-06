package main

import (
	"fmt"
	"github.com/kballard/go-shellquote"
	"os"
	"os/exec"
	"strings"
)

type Terminal interface {
	Exec(cmd string) (out []byte, err error)
	SetEnv(name, value string)
}

type LocalTerminal struct {
	env map[string]string
}

func (t *LocalTerminal) Exec(cmd string) (out []byte, err error) {
	args, err := shellquote.Split(cmd)

	if err != nil {
		return
	}

	command := exec.Command(args[0], args[1:]...)
	command.Env = t.getEnv()
	fmt.Println(command.Env)

	return command.CombinedOutput()
}

func (t *LocalTerminal) SetEnv(name, value string) {
	if t.env == nil {
		t.env = make(map[string]string, 0)
	}

	t.env[name] = value
}

func (t *LocalTerminal) getEnv() []string {
	env := make([]string, 0, 5)

	for _, v := range os.Environ() {
		parts := strings.Split(v, "=")

		if n, ok := t.env[parts[0]]; ok {
			env = append(env, parts[0]+"="+n)
		} else {
			env = append(env, v)
		}

	}

	return env
}
