package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type build struct {
	Host       string
	User       string
	Jail       string
	CgoCFlags  string `json:"CGO_CFLAGS"`
	CgoLdFlags string `json:"CGO_LDFLAGS"`
	GoRoot     string
	GoPath     string

	AppRepo    string
	AppPath    string
	RepoType   string
	RepoBranch string

	// TODO:
	PreBuildCommands  []string
	PostBuildCommands []string

	sshClient *SSHConn
	scm       Scm
}

func Build() {
	f, err := ioutil.ReadFile("./build.cfg")

	if err != nil {
		fmt.Println("Failed to read build.cfg")
		return
	}

	b := new(build)

	err = json.Unmarshal(f, b)

	if err != nil {
		fmt.Println("Failed to parse build.cfg: " + err.Error())
	}

	b.sshClient = new(SSHConn)
	b.sshClient.Connect(b.Host, b.User)
	defer b.sshClient.Close()

	b.perform()
}

func (b *build) perform() {
	b.setupScm()
	b.validateEnvironment()
	b.updateCode()
}

// Ensure all directories exist
func (b *build) validateEnvironment() {
	// Validate Jail exists
	_, err := b.sshClient.Exec("ls " + b.Jail)
	if err != nil {
		fmt.Println("Could not find Jail directory: " + err.Error())
	}

	// Validate GOROOT exists
	_, err = b.sshClient.Exec("ls " + b.GoRoot)
	if err != nil {
		fmt.Println("Could not find GOROOT directory: " + err.Error())
	}

	// Validate Go Binary exists
	_, err = b.sshClient.Exec("ls " + b.GoRoot + "/bin/go")
	if err != nil {
		fmt.Println("Could not find Go binary: " + err.Error())
	}

	// Validate Git exists
	_, err = b.sshClient.Exec("which " + b.scm.BinaryName())
	if err != nil {
		fmt.Println("Could not find " + b.RepoType + " binary: " + err.Error())
	}
}

// Checkout project from repository
func (b *build) updateCode() {
	// Fetch code base
	b.scm.SetSSHConn(b.sshClient)
	b.scm.Checkout(b.AppRepo, b.RepoBranch, b.Jail)
}

func (b *build) setupScm() {
	switch b.RepoType {
	case "git":
		b.scm = new(GitScm)

	default:
		panic("unkown RepoType")
	}
}
