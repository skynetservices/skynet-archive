package skynet

import (
	"encoding/json"
	"path"
	"strconv"
	"time"
)

type ServiceStatistics struct {
	Clients        int32
	StartTime      string
	LastRequest    string
	RequestsServed int64

	// For now this will be since startup, we might change it later to be for a given sample interval
	AverageResponseTime time.Duration
	TotalDuration       time.Duration `json:"-"`
}

type ServiceInfo struct {
	Config     *ServiceConfig
	Registered bool
	Stats      ServiceStatistics `json:"-"`
}

func (s *ServiceInfo) GetConfigPath() string {
	//	return "/services/" + s.Config.Name + "/" + s.Config.Version + "/" + s.Config.Region + "/" + s.Config.ServiceAddr.IPAddress + "/" + strconv.Itoa(s.Config.ServiceAddr.Port)
	return path.Join("/services", s.Config.Name, s.Config.Version, s.Config.Region, s.Config.ServiceAddr.IPAddress, strconv.Itoa(s.Config.ServiceAddr.Port))
}

func (s *ServiceInfo) GetStatsPath() string {
	return path.Join("/statistics", s.Config.Name, s.Config.Version, s.Config.Region, s.Config.ServiceAddr.IPAddress, strconv.Itoa(s.Config.ServiceAddr.Port))
}

func (s *ServiceInfo) FetchStats(doozer *DoozerConnection) (err error) {
	rev := doozer.GetCurrentRevision()
	data, _, err := doozer.Get(s.GetStatsPath(), rev)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, s)
	if err != nil {
		return
	}
	return
}
