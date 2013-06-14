package daemon

type SubServiceInfo struct {
	UUID        string
	ServicePath string
	Args        string
	Running     bool
}

type ListSubServicesRequest struct {
}

type ListSubServicesResponse struct {
	Services map[string]SubServiceInfo
}

type StopAllSubServicesRequest struct {
}

type StopAllSubServicesResponse struct {
	Count int
	Stops []StopSubServiceResponse
}

type StartSubServiceRequest struct {
	BinaryName string
	Args       string
}

type StartSubServiceResponse struct {
	Ok   bool
	UUID string
}

type StopSubServiceRequest struct {
	UUID string
}

type StopSubServiceResponse struct {
	Ok   bool
	UUID string
}

type RestartSubServiceRequest struct {
	UUID string
}

type RestartSubServiceResponse struct {
	UUID string
}

type RestartAllSubServicesRequest struct {
}

type RestartAllSubServicesResponse struct {
	Count    int
	Restarts []RestartSubServiceResponse
}

type RegisterSubServiceRequest struct {
	UUID string
}

type RegisterSubServiceResponse struct {
	Ok   bool
	UUID string
}

type UnregisterSubServiceRequest struct {
	UUID string
}

type UnregisterSubServiceResponse struct {
	Ok   bool
	UUID string
}

type SubServiceLogLevelRequest struct {
	UUID  string
	Level string
}

type SubServiceLogLevelResponse struct {
	Ok    bool
	UUID  string
	Level string
}

type LogLevelRequest struct {
	Level string
}

type LogLevelResponse struct {
	Ok    bool
	Level string
}

type StopRequest struct {
}

type StopResponse struct {
	Ok bool
}
