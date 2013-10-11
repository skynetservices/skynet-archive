package skynet

// ServiceHandshake is data sent by the service to the client immediately once the connection
// is opened.
type ServiceHandshake struct {
	// Name indicates the service name, for validation on the client side
	Name string

	// Registered indicates the state of this service. If it is false, the connection will
	// close immediately and the client should look elsewhere for this service.
	Registered bool

	// ClientID is a UUID that is used by the client to identify itself in RPC requests.
	ClientID string
}

// ClientHandshake is sent by the client to the service after receipt of the ServiceHandshake.
type ClientHandshake struct {
	ClientID string
}
