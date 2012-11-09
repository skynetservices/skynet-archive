package main

import (
	"bufio"
	"errors"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/daemon"
	"github.com/bketelsen/skynet/service"
	"io"
	//"log"
	"os"
	"strings"
)

// Daemon will run and maintain skynet services.
//
// Daemon will initially deploy those specified in the file given in
// the "-config" option
//
// Daemon will run the "SkynetDeployment" service, which can be used
// to remotely spawn new services on the host.
func main() {
	config, args := skynet.GetServiceConfig()

	config.Name = "SkynetDaemon"
	config.Version = "1"

	// skydaemon does not listen to admin RPC requests
	config.AdminAddr = nil

	var err error
	mlogger, err := skynet.NewMongoSemanticLogger(config.MongoConfig.MongoHosts, "skynet",
		"log", config.UUID)
	clogger := skynet.NewConsoleSemanticLogger("skydaemon", os.Stdout)
	config.Log = skynet.NewMultiSemanticLogger(mlogger, clogger)
	if err != nil {
		config.Log.Trace("Could not connect to mongo db for logging")
	}

	deployment := &SkynetDaemon{
		Log:      config.Log,
		Services: map[string]*SubService{},
	}

	s := service.CreateService(deployment, config)

	deployment.Service = s

	// handle panic so that we remove ourselves from the pool in case of catastrophic failure
	/*defer func() {
		s.Shutdown()
		if err := recover(); err != nil {
			log.Println("Unrecovered error occured: ", err)
		}
	}()*/

	if len(args) == 1 {
		err := deployConfig(deployment, args[0])
		if err != nil {
			config.Log.Error(err.Error())
		}
	}

	// If we pass false here service will not be Registered
	// we could do other work/tasks by implementing the Started method and calling Register() when we're ready
	s.Start(true).Wait()
}

// deploy each of the services listed in the provided file
func deployConfig(s *SkynetDaemon, cfg string) (err error) {
	cfgFile, err := os.Open(cfg)
	if err != nil {
		return
	}
	br := bufio.NewReader(cfgFile)
	for {
		var bline []byte
		var prefix bool
		bline, prefix, err = br.ReadLine()
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			return
		}
		if prefix {
			err = errors.New("Config line to long in " + cfg)
			return
		}
		line := strings.TrimSpace(string(bline))
		if len(line) == 0 {
			continue
		}

		split := strings.Index(line, " ")
		if split == -1 {
			split = len(line)
		}
		servicePath := line[:split]
		args := strings.TrimSpace(line[split:])
		s.Deploy(&skynet.RequestInfo{}, daemon.DeployRequest{ServicePath: servicePath, Args: args}, &daemon.DeployResponse{})
	}
	return
}
