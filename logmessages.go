package skynet

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

type MongoConnected struct {
	Addr string
}

func (m MongoConnected) String() string {
	return fmt.Sprintf("Connected to mongodb at %s", m.Addr)
}

type MongoError struct {
	Addr, Err string
}

func (m MongoError) String() string {
	return fmt.Sprintf("MongoDB error: %s: %s", m.Addr, m.Err)
}

type LogsearchClient struct {
	RemoteAddr, Method, Path string
}

func (l LogsearchClient) String() string {
	return fmt.Sprintf("Log Search client attached: %s → %s %s", l.RemoteAddr, l.Method, l.Path)
}
