package main

import (
	"path"
	"github.com/bketelsen/skynet/skylib"
	"go/build"
	"sync"
	"os/exec"
	"os"
	"path/filepath"
)

type SubService struct {
	// ServicePath is the gopath repr of the service binary
	ServicePath string
	// Args is the unprocessed command line arguments tacked on
	// after the binary name.
	Args string

	running bool
	binPath string

	rerunChan chan bool

	startMutex sync.Mutex
}

func NewSubService(log skylib.Logger, servicePath, args string) (ss *SubService, err error) {
	ss = &SubService{
		ServicePath: servicePath,
		Args:        args,
	}

	//verify that it exists on the local system

	pkg, err := build.Import(ss.ServicePath, "", 0)
	if err != nil {
		return
	}

	if pkg.Name != "main" {
		return
	}

	_, binName := path.Split(ss.ServicePath)
	binPath := filepath.Join(pkg.BinDir, binName)
	ss.binPath = binPath

	return
}

func (ss *SubService) Register() {

}

func (ss *SubService) Deregister() {

}

func (ss *SubService) Stop() {
	ss.startMutex.Lock()
	defer ss.startMutex.Unlock()

	if !ss.running {
		return
	}

	ss.Deregister()
	// halt the rerunner so we can kill the processes without it relaunching
	ss.rerunChan <- false
}

func (ss *SubService) Start() {
	ss.startMutex.Lock()
	defer ss.startMutex.Unlock()

	if ss.running {
		return
	}
	ss.rerunChan = make(chan bool)

	go ss.rerunner(ss.rerunChan)
	// send a signal to launch the service
	ss.rerunChan <- true
}

func (ss *SubService) Restart() {
	ss.Stop()
	ss.Start()
}

func (ss *SubService) rerunner(rerunChan chan bool) {
	var proc *os.Process
	for rerun := range rerunChan {
		if !rerun {
			break
		}

		cmd := exec.Command(ss.binPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Start()
		proc = cmd.Process

		// In another goroutine, wait for the process to complete and send a relaunch signal.
		// If this signal is sent after the stop signal, it is ignored.
		go func(proc *os.Process) {
			proc.Wait()
			rerunChan <- true
		}(proc)
	}
	proc.Kill()
}
