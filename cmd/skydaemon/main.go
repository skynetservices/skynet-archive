package main

import (
	"github.com/skynetservices/skynet"
	"github.com/skynetservices/skynet/log"
	"github.com/skynetservices/skynet/service"
	_ "github.com/skynetservices/zkmanager"
	"time"
)

// Daemon will run and maintain skynet services.
//
// Daemon will run the "SkynetDeployment" service, which can be used
// to remotely spawn new services on the host.
func main() {
	si := skynet.NewServiceInfo("SkynetDaemon", "2.0.0")
	deployment := NewSkynetDaemon()

	s := service.CreateService(deployment, si)

	deployment.Service = s

	// handle panic so that we remove ourselves from the pool in case of catastrophic failure
	defer func() {
		s.Shutdown()
		deployment.closeStateFile()

		if err := recover(); err != nil {
			e := err.(error)
			log.Fatal("Unrecovered error occured: " + e.Error())
		}
	}()

	// Collect Host metrics
	statTicker := time.Tick((5 * time.Second))
	go func() {
		for _ = range statTicker {
			deployment.updateHostStats(si.ServiceAddr.IPAddress)
		}
	}()

	// If we pass false here service will not be Registered
	// we could do other work/tasks by implementing the Started method and calling Register() when we're ready
	s.Start().Wait()
}
