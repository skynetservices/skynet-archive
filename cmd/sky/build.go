package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
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

	term        Terminal
	scm         Scm
	projectPath string
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

	if b.Host == "localhost" || b.Host == "127.0.0.1" || b.Host == "" {
		b.term = new(LocalTerminal)
	} else {
		sshClient := new(SSHConn)
		b.term = sshClient
		sshClient.Connect(b.Host, b.User)
		defer sshClient.Close()
	}

	b.perform()
}

func (b *build) perform() {
	b.setupScm()

	if b.validateEnvironment() {
		b.updateCode()
	}
}

// Ensure all directories exist
func (b *build) validateEnvironment() (valid bool) {
	valid = true

	// Validate Jail exists
	_, err := b.term.Exec("ls " + b.Jail)
	if err != nil {
		fmt.Println("Could not find Jail directory: " + err.Error())
		valid = false
	}

	// Validate GOROOT exists
	_, err = b.term.Exec("ls " + b.GoRoot)
	if err != nil {
		fmt.Println("Could not find GOROOT directory: " + err.Error())
		valid = false
	}

	// Validate Go Binary exists
	_, err = b.term.Exec("ls " + b.GoRoot + "/bin/go")
	if err != nil {
		fmt.Println("Could not find Go binary: " + err.Error())
		valid = false
	}

	// Validate Git exists
	_, err = b.term.Exec("which " + b.scm.BinaryName())
	if err != nil {
		fmt.Println("Could not find " + b.RepoType + " binary: " + err.Error())
		valid = false
	}

	return
}

// Checkout project from repository
func (b *build) updateCode() {
	p, err := b.scm.ImportPathFromRepo(b.AppRepo)
	b.projectPath = path.Join(b.Jail, "src", p)

	if err != nil {
		panic(err.Error())
	}

	out, err := b.term.Exec("ls " + b.projectPath)

	if err != nil {
		fmt.Println("Creating project directories")
		out, err = b.term.Exec("mkdir -p " + b.projectPath)

		if err != nil {
			panic("Could not create project directories")
		}

		fmt.Println(string(out))
	}

	// Fetch code base
	b.scm.SetTerminal(b.term)
	b.scm.Checkout(b.AppRepo, b.RepoBranch, b.projectPath)
}

func (b *build) setupScm() {
	switch b.RepoType {
	case "git":
		b.scm = new(GitScm)

	default:
		panic("unkown RepoType")
	}
}
