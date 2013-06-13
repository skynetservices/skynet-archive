package daemon

type StartRequest struct {
	BinaryName string
	Args       string
}

type StartResponse struct {
	UUID string
}

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

type StartAllSubServicesRequest struct {
}

type StartAllSubServicesResponse struct {
	Count  int
	Starts []StartSubServiceResponse
}

type StartSubServiceRequest struct {
	UUID string
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
