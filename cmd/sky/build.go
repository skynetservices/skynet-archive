package main

import (
	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/gopass"
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

	sshClient *ssh.ClientConn
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

	b.connect()
	defer b.sshClient.Close()

	b.perform()
}

func (b *build) perform() {
	b.validate()

	out, err := b.sshExec("echo $GOPATH")
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(string(out))
}

func (b *build) validate() {
	// Validate Jail exists
	out, err := b.sshExec("ls " + b.Jail)
	if err != nil {
		fmt.Println("Could not find Jail directory: " + err.Error())
	}

	// Validate GOROOT exists
	out, err = b.sshExec("ls " + b.GoRoot)
	if err != nil {
		fmt.Println("Could not find GOROOT directory: " + err.Error())
	}

	// Validate Go Binary exists
	out, err = b.sshExec("ls " + b.GoRoot + "/bin/go")
	if err != nil {
		fmt.Println("Could not find Go binary: " + err.Error())
	}

	// Validate Git exists
	out, err = b.sshExec("which git")
	if err != nil {
		fmt.Println("Could not find Git binary: " + err.Error())
	}
}

func (b build) Password(user string) (string, error) {
	pass, err := gopass.GetPass("Password for " + user + ": ")

	if err != nil {
		panic("Failed to collect password: " + err.Error())
	}

	return pass, err
}

// SSH logic
func (b *build) connect() {
	config := &ssh.ClientConfig{
		User: b.User,
		Auth: []ssh.ClientAuth{
			ssh.ClientAuthPassword(b),
		},
	}

	var err error
	b.sshClient, err = ssh.Dial("tcp", b.Host, config)
	if err != nil {
		panic("Failed to dial: " + err.Error())
	}
}

func (b *build) sshExec(cmd string) (out []byte, err error) {
	var session *ssh.Session

	session, err = b.sshClient.NewSession()
	if err != nil {
		panic("Failed to create session: " + err.Error())
	}
	defer session.Close()

	return session.CombinedOutput(cmd)
}
