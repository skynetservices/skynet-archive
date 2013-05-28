package main

import (
	"fmt"
	"github.com/skynetservices/go-shellquote"
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/client"
	"os"
	"os/exec"
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

	runSignal sync.WaitGroup

	UUID string

	doozer *skynet.DoozerConnection
}

func NewSubService(daemon *SkynetDaemon, servicePath, args, uuid string) (ss *SubService, err error) {
	ss = &SubService{
		ServicePath: servicePath,
		Args:        args,
		UUID:        uuid,
		doozer:      daemon.Service.DoozerConn,
		// TODO: proper argument splitting
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
	ss.binPath = filepath.Join(bindir, servicePath)

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

	if !ss.running {
		return false
	}
	ss.running = false

	ss.Deregister()

	// halt the rerunner so we can kill the processes without it relaunching
	ss.runSignal.Add(1)
	ss.rerunChan <- false
	ss.runSignal.Wait()

	ss.sendRPCStop()

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
	// Because we don't call stop here,
	// rerunner will spawn a new instance when the old one has stopped
	ss.sendRPCStop()
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

	ss.removeFromDoozer()

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

func (ss *SubService) removeFromDoozer() {
	q := skynet.Query{
		UUID:       ss.UUID,
		DoozerConn: ss.doozer,
	}

	instances := q.FindInstances()
	for _, instance := range instances {
		ss.doozer.Del(instance.GetConfigPath(), ss.doozer.GetCurrentRevision())
	}
}

func (ss *SubService) sendRPCStop() {
	q := skynet.Query{
		UUID:       ss.UUID,
		DoozerConn: ss.doozer,
	}
	instances := q.FindInstances()
	for _, instance := range instances {
		cladmin := client.Admin{
			Instance: instance,
		}

		cladmin.Stop(skynet.StopRequest{
			WaitForClients: true,
		})
	}

}
