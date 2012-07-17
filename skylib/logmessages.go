package skylib

import (
	"fmt"
)

type DoozerConnected struct {
	Addr string
}

func (dc DoozerConnected) String() string {
	return fmt.Sprintf("Connected to doozer at %s", dc.Addr)
}

type DoozerDiscovered struct {
	DoozerServer *DoozerServer
}

func (dd DoozerDiscovered) String() string {
	return fmt.Sprintf("Discovered new doozer %s at %s", dd.DoozerServer.Key, dd.DoozerServer.Addr)
}

type DoozerRemoved struct {
	DoozerServer *DoozerServer
}

func (dr DoozerRemoved) String() string {
	return fmt.Sprintf("Removed doozer %s at %s", dr.DoozerServer.Key, dr.DoozerServer.Addr)
}

type DoozerLostConnection struct {
	DoozerConfig *DoozerConfig
}

func (dlc DoozerLostConnection) String() string {
	return fmt.Sprintf("Lost connection to doozer at %s", dlc.DoozerConfig.Uri)
}

type ServiceDiscovered struct {
	Service *Service
}

func (sd ServiceDiscovered) String() string {
	return fmt.Sprintf("Discovered service %q at %s", sd.Service.Config.Name, sd.Service.Config.ServiceAddr)
}

type ServiceRemoved struct {
	Service *Service
}

func (sr ServiceRemoved) String() string {
	return fmt.Sprintf("Removed service %q at %s", sr.Service.Config.Name, sr.Service.Config.ServiceAddr)
}

type ServiceCreated struct {
	ServiceConfig *ServiceConfig
}

func (sc ServiceCreated) String() string {
	return fmt.Sprintf("Created service %q", sc.ServiceConfig.Name)
}

type ServiceListening struct {
	ServiceConfig *ServiceConfig
	Addr          *BindAddr
}

func (sc ServiceListening) String() string {
	return fmt.Sprintf("Service %q listening on %s", sc.ServiceConfig.Name, sc.Addr)
}

type AdminListening struct {
	ServiceConfig *ServiceConfig
}

func (al AdminListening) String() string {
	return fmt.Sprintf("Service %q listening for admin on %s", al.ServiceConfig.Name, al.ServiceConfig.AdminAddr)
}
