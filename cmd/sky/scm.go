package main

import (
	"fmt"
)

type Scm interface {
	SetSSHConn(sshClient *SSHConn)
	Checkout(repo, branch, path string) error
	BinaryName() string
}

type GitScm struct {
	sshClient *SSHConn
}

func (g *GitScm) BinaryName() string {
	return "git"
}

func (g *GitScm) SetSSHConn(sshClient *SSHConn) {
	g.sshClient = sshClient
}

func (g *GitScm) Checkout(repo, branch, path string) (err error) {
	out, err := g.sshClient.Exec("ls " + path)

	if err != nil {
		fmt.Println(err.Error())
	}

	// If repo doesn't exist we need to clone it
	if len(out) == 0 {
		fmt.Println("Checkout out repo: " + repo)
		out, err = g.sshClient.Exec("git clone " + repo + " " + path)
		fmt.Println(string(out))

		if err != nil {
			return
		}
	} else {
		// Repo exists, get latest and checkout correct branch
		fmt.Println("Fetching latest from repo: " + repo)
		out, err = g.sshClient.Exec("cd " + path + "&& git pull")
		fmt.Println(string(out))

		if err != nil {
			return
		}
	}

	// Ensure we are on the correct branch
	fmt.Println("Checkout out branch: " + branch)
	out, err = g.sshClient.Exec("cd " + path + "&& git checkout " + branch)
	fmt.Println(string(out))

	if err != nil {
		return
	}

	return err
}
