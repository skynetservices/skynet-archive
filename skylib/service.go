//Copyright (c) 2011 Brian Ketelsen

//Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package skylib

import (
	"os"
	"time"
)

// A Generic struct to represent any service in the SkyNet system.
type Service struct {
	IPAddress string
	Name      string
	Port      int
	Provides  string
}

// Exported RPC method for the health check
func (hc *Service) Ping(hr *HeartbeatRequest, resp *HeartbeatResponse) (err os.Error) {

	resp.Timestamp = time.Seconds()
	resp.Ok = true

	return nil
}

// Exported RPC method for the advanced health check
func (hc *Service) PingAdvanced(hr *HealthCheckRequest, resp *HealthCheckResponse) (err os.Error) {

	resp.Timestamp = time.Seconds()
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
	if r.Provides != that.Provides {
		return b
	}
	b = true
	return b
}

// Utility function to return a new Service struct
// pre-populated with the data on the command line.
func NewService(provides string) *Service {
	r := &Service{
		Name:      *Name,
		Port:      *Port,
		IPAddress: *BindIP,
		Provides:  provides,
	}

	return r
}
