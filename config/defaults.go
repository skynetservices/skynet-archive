package config

import (
	"fmt"
	"github.com/skynetservices/skynet2/log"
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
	DefaultRegion      = "unknown"
	DefaultVersion     = "unknown"
	DefaultHost        = "127.0.0.1"
	DefaultMinPort     = 9000
	DefaultMaxPort     = 9999

	DefaultLogLevel = log.DEBUG
)

func GetDefaultBindAddr() string {
	return fmt.Sprintf("%s:%d-%d", DefaultHost, DefaultMinPort, DefaultMaxPort)
}
