package skynet

import (
	"encoding/json"
	"path"
	"strconv"
	"time"
)

// ServiceStatistics contains information about its service that can
// be used to estimate load.
type ServiceStatistics struct {
	// Clients is the number of clients currently connected to this service.
	Clients int32
	// StartTime is the time when the service began running.
	StartTime string
	// LastRequest is the time when the last request was made.
	LastRequest string
	// RequestsServed is the number of requests served by this service
	// since it began.
	RequestsServed int64

	// AverageResponseTime is the average time taken to respond to a
	// request, since startup.
	// Note: in the future, this may be the average over some sliding window.
	AverageResponseTime time.Duration

	// TotalDuration is the total time taken by all requests made to
	// this service.
	TotalDuration time.Duration `json:"-"`
}

// ServiceInfo is the publicly reported information about a particular
// service instance.
type ServiceInfo struct {
	// Config is the configuration used to start this instance.
	Config *ServiceConfig
	// Registered indicates if the instance is currently accepting requests.
	Registered bool
	// Stats is transient data about instance load and other things.
	Stats *ServiceStatistics `json:",omitempty"`
}

// *ServiceInfo.GetConfigPath() returns the doozer path where it's
// stored. The statistics are not included.
func (s *ServiceInfo) GetConfigPath() string {
	return path.Join("/services", s.Config.Name, s.Config.Version,
		s.Config.Region, s.Config.ServiceAddr.IPAddress,
		strconv.Itoa(s.Config.ServiceAddr.Port))
}

// *ServiceInfo.GetStatsPath() returns the doozer path where it's
// statistics are stored.
func (s *ServiceInfo) GetStatsPath() string {
	return path.Join("/statistics", s.Config.Name, s.Config.Version,
		s.Config.Region, s.Config.ServiceAddr.IPAddress,
		strconv.Itoa(s.Config.ServiceAddr.Port))
}

// *ServiceInfo.FetchStats will query the provided doozer connection
// and update its .Stats field.
func (s *ServiceInfo) FetchStats(doozer *DoozerConnection) (err error) {
	rev := doozer.GetCurrentRevision()
	data, _, err := doozer.Get(s.GetStatsPath(), rev)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &s.Stats)
	if err != nil {
		return
	}
	return
}
