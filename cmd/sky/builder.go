package main

import (
	"encoding/json"
	"fmt"
	"go/build"
	"io/ioutil"
	"path"
	"strings"
)

type builder struct {
	BuildConfig  buildConfig `json:"Build"`
	DeployConfig buildConfig `json:"Deploy"`

	term        Terminal
	scm         Scm
	projectPath string
	pack        *build.Package
}

type buildConfig struct {
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

	UpdatePackages   bool
	BuildAllPackages bool
	RunTests         bool

	// TODO:
	PreBuildCommands  []string
	PostBuildCommands []string
}

type deployConfig struct {
	DeployPath string
}

var context = build.Default

func Build() {
	f, err := ioutil.ReadFile("./build.cfg")

	if err != nil {
		fmt.Println("Failed to read build.cfg")
		return
	}

	b := new(builder)

	err = json.Unmarshal(f, b)

	if err != nil {
		fmt.Println("Failed to parse build.cfg: " + err.Error())
	}

	if b.BuildConfig.Host == "localhost" || b.BuildConfig.Host == "127.0.0.1" || b.BuildConfig.Host == "" {
		b.term = new(LocalTerminal)
	} else {
		sshClient := new(SSHConn)
		b.term = sshClient
		sshClient.Connect(b.BuildConfig.Host, b.BuildConfig.User)
		defer sshClient.Close()
	}

	b.perform()
}

func (b *builder) perform() {
	b.setupScm()

	if b.validateEnvironment() {
		b.updateCode()

		b.term.SetEnv("GOPATH", b.goPath())
		b.term.SetEnv("GOROOT", b.BuildConfig.GoRoot)
		b.term.SetEnv("CGO_CFLAGS", b.BuildConfig.CgoCFlags)
		b.term.SetEnv("CGO_LDFLAGS", b.BuildConfig.CgoLdFlags)
		b.fetchDependencies()
	}
}

// Ensure all directories exist
func (b *builder) validateEnvironment() (valid bool) {
	valid = true

	// Validate this package is a command
	var err error
	b.pack, err = context.ImportDir(".", 0)

	if err != nil {
		panic("Could not import package for validation")
	}

	if !b.pack.IsCommand() {
		panic("Package is not a command")
	}

	// Validate Jail exists
	_, err = b.term.Exec("ls " + b.BuildConfig.Jail)
	if err != nil {
		fmt.Println("Could not find Jail directory: " + err.Error())
		valid = false
	}

	// Validate GOROOT exists
	_, err = b.term.Exec("ls " + b.BuildConfig.GoRoot)
	if err != nil {
		fmt.Println("Could not find GOROOT directory: " + err.Error())
		valid = false
	}

	// Validate Go Binary exists
	_, err = b.term.Exec("ls " + b.BuildConfig.GoRoot + "/bin/go")
	if err != nil {
		fmt.Println("Could not find Go binary: " + err.Error())
		valid = false
	}

	// Validate Git exists
	_, err = b.term.Exec("which " + b.scm.BinaryName())
	if err != nil {
		fmt.Println("Could not find " + b.BuildConfig.RepoType + " binary: " + err.Error())
		valid = false
	}

	return
}

// Checkout project from repository
func (b *builder) updateCode() {
	p, err := b.scm.ImportPathFromRepo(b.BuildConfig.AppRepo)
	b.projectPath = path.Join(b.BuildConfig.Jail, "src", p)

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
	b.scm.Checkout(b.BuildConfig.AppRepo, b.BuildConfig.RepoBranch, b.projectPath)
}

func (b *builder) setupScm() {
	switch b.BuildConfig.RepoType {
	case "git":
		b.scm = new(GitScm)

	default:
		panic("unkown RepoType")
	}
}

func (b *builder) fetchDependencies() {
	flags := []string{"-d"}

	if b.BuildConfig.UpdatePackages {
		flags = append(flags, "-u")
	}

	for _, i := range b.pack.Imports {
		if !build.IsLocalImport(i) {

			// Skip packages that were pulled down as part of repo
			if strings.HasPrefix(path.Join(b.BuildConfig.Jail, "src", i), b.projectPath) {
				continue
			}

			fmt.Println("Fetching package: " + i)
			out, err := b.term.Exec("go get " + strings.Join(flags, " ") + " " + i)
			fmt.Println(string(out))

			if err != nil {
				panic("Failed to fetch dependency: " + i + "\n" + err.Error())
			}
		}
	}
}

func (b *builder) goPath() string {
	// TODO: we need a clean gopath in case it imports itself, then we also need to append ourself so the package can be found here?
	// another issue is if it fetches dependencies it could try to fetch itself, might need to change the gopath again once we've done the update?
	return "/Users/erikstmartin/tmp" //b.BuildConfig.Jail + ":" + b.BuildConfig.GoPath
}
