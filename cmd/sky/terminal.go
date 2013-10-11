package main

import (
	"github.com/kballard/go-shellquote"
	"os"
	"os/exec"
	"strings"
)

type Terminal interface {
	Exec(cmd string) (out []byte, err error)
	ExecPath(cmd, path string) (out []byte, err error)
	SetEnv(name, value string)
	Close()
}

type LocalTerminal struct {
	env map[string]string
}

func (t *LocalTerminal) ExecPath(cmd, path string) (out []byte, err error) {
	args, err := shellquote.Split(cmd)

	if err != nil {
		return
	}

	command := exec.Command(args[0], args[1:]...)
	command.Env = t.getEnv()

	if path != "" {
		command.Dir = path
	}

	return command.CombinedOutput()
}

func (t *LocalTerminal) SetEnv(name, value string) {
	if t.env == nil {
		t.env = make(map[string]string, 0)
	}

	t.env[name] = value
}

func (t *LocalTerminal) Exec(cmd string) (out []byte, err error) {
	return t.ExecPath(cmd, "")
}

func (t *LocalTerminal) getEnv() []string {
	env := make([]string, 0, 5)

	for _, v := range os.Environ() {
		parts := strings.Split(v, "=")

		if _, ok := t.env[parts[0]]; !ok {
			env = append(env, v)
		}
	}

	for k, v := range t.env {
		env = append(env, k+"="+v)
	}

	return env
}

func (t *LocalTerminal) Close() {
	// noop
}
