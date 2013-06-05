package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
)

type build struct {
	Host       string
	Jail       string
	CgoCFlags  string `json:"CGO_CFLAGS"`
	CgoLdFlags string `json:"CGO_LDFLAGS"`
	GoRoot     string
	GoPath     string

	AppRepo    string
	AppPath    string
	RepoType   string
	RepoBranch string

	PreBuildCommands  []string
	PostBuildCommands []string
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

	b.Perform()
}

func (b *build) Perform() {
	b.Validate()
}

func (b *build) Validate() {
	cmd := exec.Command("ssh", b.Host, "'ls "+b.Jail+"'")
	fmt.Println(cmd.Args)
	err := cmd.Run()

	if err != nil {
		fmt.Println("SSH Error: " + err.Error())
		return
	}

	fmt.Println("success")
}
