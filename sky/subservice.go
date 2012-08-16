package main

import (
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

	running bool
	binPath string

	rerunChan chan bool

	startMutex sync.Mutex
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
	// TODO: connect to admin port or remove this method
}

func (ss *SubService) Deregister() {
	// TODO: connect to admin port or remove this method
}

func (ss *SubService) Stop() {
	ss.startMutex.Lock()
	defer ss.startMutex.Unlock()

	if !ss.running {
		return
	}
	ss.running = false

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
	ss.running = true
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

		cmd := exec.Command(ss.binPath, ss.argv...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		startupTimer := time.NewTimer(RerunWait)

		cmd.Start()
		proc = cmd.Process

		// In another goroutine, wait for the process to complete and send a relaunch signal.
		// If this signal is sent after the stop signal, it is ignored.
		go func(proc *os.Process) {
			proc.Wait()
			select {
			case <-startupTimer.C:
				// we let it run long enough that it might not be a recurring error, try again
				rerunChan <- true
			default:
				// error happened too quickly - must be startup issue
				startupTimer.Stop()
			}
		}(proc)
	}
	proc.Kill()
}
