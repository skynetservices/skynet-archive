package main

import (
	"errors"
	"fmt"
	"github.com/bketelsen/skynet"
	"github.com/kballard/go-shellquote"
	"go/build"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sync"
	"time"
)

const RerunWait = time.Second * 5

type SubService struct {
	// ServicePath is the gopath repr of the service binary
	ServicePath string
	// Args is the unprocessed command line arguments tacked on
	// after the binary name.
	Args string
	// argv is Args after it is properly split up
	argv []string

	running      bool
	runningMutex sync.Mutex

	binPath string

	rerunChan chan bool

	startMutex sync.Mutex

	runSignal sync.WaitGroup
}

func NewSubService(log skynet.Logger, servicePath, args, uuid string) (ss *SubService, err error) {
	ss = &SubService{
		ServicePath: servicePath,
		Args:        args,
		// TODO: proper argument splitting
	}
	ss.argv, err = shellquote.Split(args)
	if err != nil {
		return
	}

	ss.argv = append([]string{"-uuid", uuid}, ss.argv...)

	// verify that it exists on the local system
	// TODO: go get package?
	pkg, err := build.Import(ss.ServicePath, "", 0)
	if err != nil {
		return
	}

	if pkg.Name != "main" {
		return nil, errors.New("This package is not a binary")
	}

	_, binName := path.Split(ss.ServicePath)
	binPath := filepath.Join(pkg.BinDir, binName)

	ss.binPath = binPath

	return
}

func (ss *SubService) Register() {
	// TODO: connect to admin port or remove this method
}

func (ss *SubService) Deregister() {
	// TODO: connect to admin port or remove this method
}

func (ss *SubService) Stop() bool {
	ss.startMutex.Lock()
	defer ss.startMutex.Unlock()

	ss.runningMutex.Lock()
	defer ss.runningMutex.Unlock()

	if !ss.running {
		return false
	}
	ss.running = false

	ss.Deregister()
	// halt the rerunner so we can kill the processes without it relaunching
	ss.runSignal.Add(1)

	ss.rerunChan <- false

	ss.runSignal.Wait()

	return true
}

func (ss *SubService) Start() (success bool, err error) {
	success = false

	ss.startMutex.Lock()
	defer ss.startMutex.Unlock()

	ss.runningMutex.Lock()
	defer ss.runningMutex.Unlock()

	if ss.running {
		return
	}
	ss.running = true
	ss.rerunChan = make(chan bool)
	go ss.rerunner()

	// Block for first start so we can make sure binary exists etc
	_, err = ss.startProcess()

	if err == nil {
		success = true
	}

	return
}

func (ss *SubService) Restart() {
	ss.Stop()
	ss.Start()
}

func (ss *SubService) startProcess() (proc *os.Process, err error) {
	cmd := exec.Command(ss.binPath, ss.argv...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	proc = cmd.Process
	startupTimer := time.NewTimer(RerunWait)

	if proc != nil {
		go ss.watchProcess(proc, startupTimer)
	}

	return
}

func (ss *SubService) watchProcess(proc *os.Process, startupTimer *time.Timer) {
	proc.Wait()

	ss.runningMutex.Lock()
	defer ss.runningMutex.Unlock()

	if !ss.running {
		startupTimer.Stop()
		return
	}

	select {
	case <-startupTimer.C:
		// we let it run long enough that it might not be a recurring error, try again
		if ss.running {
			ss.rerunChan <- true
		}
	default:
		// error happened too quickly - must be startup issue
		startupTimer.Stop()
	}
}

func (ss *SubService) rerunner() {
	for rerun := range ss.rerunChan {

		if !rerun {
			break
		}

		fmt.Println("Restarting SubService: " + ss.binPath)
		_, err := ss.startProcess()

		if err != nil {
			fmt.Println(err)
		}
	}

	ss.runSignal.Done()
}
