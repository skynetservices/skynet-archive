package main

import (
	"fmt"
	"github.com/kballard/go-shellquote"
	"github.com/skynetservices/skynet2/daemon"
	"github.com/skynetservices/skynet2/log"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
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

	runSignal sync.WaitGroup

	UUID string

	pipe *daemon.Pipe
}

func NewSubService(binaryName, args, uuid string) (ss *SubService, err error) {
	ss = &SubService{
		ServicePath: binaryName,
		Args:        args,
		UUID:        uuid,
	}

	ss.argv, err = shellquote.Split(args)
	if err != nil {
		return
	}

	ss.argv = append([]string{"-uuid", uuid}, ss.argv...)

	bindir := os.Getenv("SKYNET_SERVICE_DIR")
	if bindir == "" {
		bindir = "/usr/bin"
	}
	ss.binPath = filepath.Join(bindir, binaryName)

	return
}

func (ss *SubService) Register() bool {
	// TODO: connect to admin port or remove this method
	return true
}

func (ss *SubService) Unregister() bool {
	// TODO: connect to admin port or remove this method
	return true
}

func (ss *SubService) Stop() bool {
	ss.startMutex.Lock()
	defer ss.startMutex.Unlock()

	if !ss.running {
		return false
	}

	ss.runSignal.Add(1)
	close(ss.rerunChan)

	ss.running = false

	ss.Unregister()

	// TODO: Clean this up and put it somewhere useful abstracted cleanly
	ss.pipe.Write([]byte("SHUTDOWN"))
	b := make([]byte, daemon.MAX_PIPE_BYTES)

	n, err := ss.pipe.Read(b)
	if err != nil && err != io.EOF {
		log.Println(log.ERROR, err)
	} else {
		log.Println(log.TRACE, "Read From Service: "+string(b[:n]))
	}

	// halt the rerunner so we can kill the processes without it relaunching
	ss.runSignal.Wait()

	return true
}

func (ss *SubService) Start() (success bool, err error) {
	success = false

	ss.startMutex.Lock()
	defer ss.startMutex.Unlock()

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
	// TODO:
}

func (ss *SubService) startProcess() (proc *os.Process, err error) {
	cmd := exec.Command(ss.binPath, ss.argv...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	dRead, ssWrite, err := os.Pipe()
	ssRead, dWrite, err := os.Pipe()

	// Store Fd's for us to read/write to SubService
	ss.pipe = daemon.NewPipe(dRead, dWrite)

	// Pass SubService their side of the connection
	cmd.ExtraFiles = append(cmd.ExtraFiles, ssRead, ssWrite)

	err = cmd.Start()
	proc = cmd.Process
	startupTimer := time.NewTimer(RerunWait)

	if proc != nil {
		go ss.watchSignals()
		go ss.watchProcess(proc, startupTimer)
	}

	return
}

func (ss *SubService) watchProcess(proc *os.Process, startupTimer *time.Timer) {
	proc.Wait()

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
		fmt.Println("Service died too quickly: " + ss.binPath)
		startupTimer.Stop()
	}
}

func (ss *SubService) rerunner() {
	for rerun := range ss.rerunChan {
		if !rerun {
			break
		}

		fmt.Println("Restarting service: " + ss.binPath)
		_, err := ss.startProcess()

		if err != nil {
			fmt.Println(err)
		}
	}

	ss.runSignal.Done()
}

func (ss *SubService) watchSignals() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGKILL, syscall.SIGSEGV, syscall.SIGSTOP, syscall.SIGTERM)

	for {
		select {
		case sig := <-c:
			switch sig.(syscall.Signal) {
			// Trap signals so we don't get restarted
			case syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT,
				syscall.SIGSEGV, syscall.SIGSTOP, syscall.SIGTERM:
				ss.Stop()
			}
		}
	}
}
