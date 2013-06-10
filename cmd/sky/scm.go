package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type Scm interface {
	SetTerminal(terminal Terminal)
	Checkout(repo, branch, path string) error
	ImportPathFromRepo(repoUrl string) (importPath string, err error)
	BinaryName() string
}

var gitImportRegex = regexp.MustCompile("(?:[^@]+@|(?:https?|git)://)(.+)\\.git")

type GitScm struct {
	term Terminal
}

func (g *GitScm) BinaryName() string {
	return "git"
}

func (g *GitScm) SetTerminal(terminal Terminal) {
	g.term = terminal
}

func (g *GitScm) Checkout(repo, branch, path string) (err error) {
	out, err := g.term.Exec("ls " + path)

	if err != nil {
		fmt.Println(err.Error())
	}

	// If repo doesn't exist we need to clone it
	if len(out) == 0 {
		fmt.Println("Checkout out repo: " + repo)
		out, err = g.term.Exec("git clone " + repo + " " + path)
		fmt.Println(string(out))

		if err != nil {
			return
		}
	} else {
		// Move to master to ensure we are on a branch
		err = g.CheckoutBranch("master", path)
		if err != nil {
			return
		}

		fmt.Println("Fetching latest from repo: " + repo)
		out, err = g.term.ExecPath("git pull", path)
		fmt.Println(string(out))

		if err != nil {
			return
		}
	}

	// Ensure we are on the correct branch
	fmt.Println("Checkout out branch: " + branch)

	err = g.CheckoutBranch(branch, path)
	if err != nil {
		return
	}

	// Check for submodules
	fmt.Println("Checking for submodules: " + branch)
	out, err = g.term.ExecPath("git submodule init", path)
	fmt.Println(string(out))

	if err != nil {
		return
	}

	out, err = g.term.ExecPath("git submodule update", path)
	fmt.Println(string(out))

	if err != nil {
		return
	}

	return
}

func (g *GitScm) ImportPathFromRepo(repoUrl string) (importPath string, err error) {
	matches := gitImportRegex.FindStringSubmatch(repoUrl)

	if matches == nil || len(matches) < 2 {
		return "", errors.New("Could not determine import path from repo url: " + repoUrl)
	}

	importPath = strings.Replace(matches[1], ":", "/", -1)

	return
}

func (g *GitScm) CheckoutBranch(branch, path string) error {
	out, err := g.term.ExecPath("git checkout "+branch, path)
	fmt.Println(string(out))

	return err
}
