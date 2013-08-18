package skynet

// RequestInfo is information about a request, and is provided to every skynet RPC call.
type RequestInfo struct {
	// ServiceName is the name of the service request is intended to be processed by
	ServiceName string
	// OriginAddress is the reported address of the originating client, typically from outside the service cluster.
	OriginAddress string
	// ConnectionAddress is the address of the TCP connection making the current RPC request.
	ConnectionAddress string
	// RequestID is a unique ID for the current RPC request.
	RequestID string
	// RetryCount indicates how many times this request has been tried before.
	RetryCount int
}
