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
	DefaultIdleConnectionsToInstance = 1
	// DefaultMaxConnectionsToInstance is the maximum number of concurrent connections to a particular instance.
	DefaultMaxConnectionsToInstance = 1
)

// skynet
const (
	// DefaultDoozerdAddr is where a skynet service or client will look for doozerd.
	DefaultDoozerdAddr = "127.0.0.1:8046"
	// DefaultIdleTimeout is how long a connection can remain inactive in the pool before being closed.
	DefaultIdleTimeout = 0
	// DefaultDoozerUpdateInterval is the minimum wait before sending new information to doozerd.
	DefaultDoozerUpdateInterval = 5 * time.Second
	// DefaultRegion is the region specified for a service.
	DefaultRegion = "unknown"
	// DefaultVersion is the version specified for a service.
	DefaultVersion = "unknown"
	//Default Ip Address and port for Statsd
	DefaultStatsdAddr = "127.0.0.1:8125"
	//Default Directory name for metrics in Graphite
	DefaultStatsdDir = "default"
)

func GetDefaultBindAddr() string {
	host := GetDefaultEnvVar("SKYNET_BIND_IP", "127.0.0.1")
	minPort := GetDefaultEnvVar("SKYNET_MIN_PORT", "9000")
	maxPort := GetDefaultEnvVar("SKYNET_MAX_PORT", "9999")

	return host + ":" + minPort + "-" + maxPort
}
