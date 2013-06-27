package skynet

import (
	"time"
)

// skynet/client
const (
	// DefaultRetryDuration is how long a client.ServiceClient waits before sending a new request.
	DefaultRetryDuration = 2 * time.Second
	// DefaultTimeoutDuration is how long a client.ServiceClient will wait before giving up.
	DefaultTimeoutDuration = 10 * time.Second
	// DefaultIdleConnectionsToInstance is the number of connections to a particular instance that may sit idle.
	DefaultIdleConnectionsToInstance = 2
	// DefaultMaxConnectionsToInstance is the maximum number of concurrent connections to a particular instance.
	DefaultMaxConnectionsToInstance = 5
)

// skynet
const (
	DefaultIdleTimeout = 0
	// DefaultRegion is the region specified for a service.
	DefaultRegion = "unknown"
	// DefaultVersion is the version specified for a service.
	DefaultVersion = "unknown"
)

func GetDefaultBindAddr() string {
	host := GetDefaultEnvVar("SKYNET_BIND_IP", "127.0.0.1")
	minPort := GetDefaultEnvVar("SKYNET_MIN_PORT", "9000")
	maxPort := GetDefaultEnvVar("SKYNET_MAX_PORT", "9999")

	return host + ":" + minPort + "-" + maxPort
}
