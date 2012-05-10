package skylib

import (
  "time"
  "strconv"
)

// A Generic struct to represent any service in the SkyNet system.
type Service struct {
	IPAddress  string
	Name       string
	Port       int
  Region     string
	Idempotent bool
	Version	  string

}

func GetServicePath(name *string, version *string, ip *string, port *int, region *string) (string){
  return "/services/" + *name + "/" + *version + "/" + *region + "/" + *ip + "/" + strconv.Itoa(*port)
}

func (r *Service) parseError(err string) {
	panic(&Error{err, r.Name})
}

// Exported RPC method for the health check
func (hc *Service) Admin(hr *AdminRequest, resp *AdminResponse) (err error) {
	if hr.Command == "SHUTDOWN" {
		gracefulShutdown()
	}

	return nil
}

// Exported RPC method for the health check
func (hc *Service) Ping(hr *HeartbeatRequest, resp *HeartbeatResponse) (err error) {
	resp.Timestamp = time.Now()

	return nil
}

// Exported RPC method for the advanced health check
func (hc *Service) PingAdvanced(hr *HealthCheckRequest, resp *HealthCheckResponse) (err error) {
	resp.Timestamp = time.Now()
	resp.Load = 0.1 //todo
	return nil
}

func (r *Service) Equal(that *Service) bool {
	var b bool
	b = false
	if r.Name != that.Name {
		return b
	}
	if r.IPAddress != that.IPAddress {
		return b
	}
	if r.Port != that.Port {
		return b
	}
	b = true
	return b
}

// Utility function to return a new Service struct
// pre-populated with the data on the command line.
func NewService(region string, provides string, idempotent bool, version string) *Service {
	return  &Service{
		Name:      provides,
		Port:      *Port,
		IPAddress: *BindIP,
		Idempotent: idempotent,
		Version: version,
    Region: region,
	}
}
