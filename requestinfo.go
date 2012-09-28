package skynet

import (
	"net"
)

type RequestInfo struct {
	// OriginAddress is the reported address of the originating client, typically from outside the service cluster.
	OriginAddress net.Addr
	// ConnectionAddress is the address of the TCP connection making the current RPC request.
	ConnectionAddress net.Addr
	// RequestID is a unique ID for the current RPC request.
	RequestID  string
	RetryCount int
}
