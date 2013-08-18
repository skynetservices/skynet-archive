package skynet

// ServiceStatistics contains information about its service that can
// be used to estimate load.
type ServiceStatistics struct {
	// Clients is the number of clients currently connected to this service.
	Clients int32
	// StartTime is the time when the service began running.
	StartTime string
	// LastRequest is the time when the last request was made.
	LastRequest string
}

// ServiceInfo is the publicly reported information about a particular
// service instance.
type ServiceInfo struct {
	// TODO: is there any reason this needs to be a pointer?
	// Config is the configuration used to start this instance.
	*ServiceConfig
	// Registered indicates if the instance is currently accepting requests.
	Registered bool
}

func NewServiceInfo(c *ServiceConfig) ServiceInfo {
	if c == nil {
		c = &ServiceConfig{
			UUID: UUID(),
		}
	}

	return ServiceInfo{
		ServiceConfig: c,
	}
}
