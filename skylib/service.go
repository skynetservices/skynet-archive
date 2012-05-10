package skylib

import "time"

// A Generic struct to represent any service in the SkyNet system.
type Service struct {
	IPAddress string
	Name      string
	Port      int
	Idempotent bool
	Version	int

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
func NewService(provides string, idempotent bool, version int) *Service {
	return  &Service{
		Name:      provides,
		Port:      *Port,
		IPAddress: *BindIP,
		Idempotent: idempotent,
		Version: version,
	}
}
