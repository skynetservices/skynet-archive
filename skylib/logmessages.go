package skylib

import ()

type ConnectedToDoozer struct {
	Addr string
}

type RegisteredMethod struct {
	Method string
}

type NewDoozerDetected struct {
	DoozerServer *DoozerServer
}

type DoozerNoLongerAvailable struct {
	DoozerServer *DoozerServer
}
