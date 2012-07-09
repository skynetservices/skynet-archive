package skylib

import ()

type DoozerConnected struct {
	Addr string
}
type DoozerDiscovered struct {
	DoozerServer *DoozerServer
}

type DoozerRemoved struct {
	DoozerServer *DoozerServer
}

type DoozerLostConnection struct {
	DoozerConfig *DoozerConfig
}

type RegisteredMethod struct {
	Method string
}

type ServiceDiscovered struct {
	Service *Service
}

type ServiceRemoved struct {
	Service *Service
}
