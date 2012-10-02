package skynet

type RequestInfo struct {
	// OriginAddress is the reported address of the originating client, typically from outside the service cluster.
	OriginAddress string
	// ConnectionAddress is the address of the TCP connection making the current RPC request.
	ConnectionAddress string
	// RequestID is a unique ID for the current RPC request.
	RequestID string
	// RetryCount indicates how many times this request has been tried before.
	RetryCount int
}
