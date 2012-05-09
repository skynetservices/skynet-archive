//Copyright (c) 2011 Brian Ketelsen

//Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package skylib

import (
	"fmt"
	"net/rpc"
	"time"
)


func (r *Service) parseError(err string) {
	panic(&Error{err, r.Name})
}

// A Generic struct to represent any service in the SkyNet system.
type Service struct {
	IPAddress string
	Name      string
	Port      int
	Idempotent bool
	Version	int

}

// A HeartbeatRequest is the struct that is sent for ping checks.
type HeartbeatRequest struct {
	Timestamp int64
}

// HeartbeatResponse is the struct that is returned on a ping check.
type HeartbeatResponse struct {
	Timestamp time.Time
	Ok        bool
}


type AdminRequest struct {
	Command string
}


type AdminResponse struct {
	Timestamp time.Time
	Ok        bool
}




// HealthCheckRequest is the struct that is sent on a more advanced heartbeat request.
type HealthCheckRequest struct {
	Timestamp time.Time
}

// HealthCheckResponse is the struct that is sent back to the HealthCheckRequest-er
type HealthCheckResponse struct {
	Timestamp time.Time
	Load      float64
}



type ServerConfig interface {
	Equal(that interface{}) bool
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

// Method to register the heartbeat of each skynet
// client with the healthcheck exporter.
func RegisterHeartbeat() {
	r := NewService("Ping", false, 1)
	rpc.Register(r)
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

type Error struct {
	Msg     string
	Service string
}

func (e *Error) Error() string { return fmt.Sprintf("Service %s had error: %s", e.Service, e.Msg) }

func NewError(msg string, service string) (err *Error) {
	err = &Error{Msg: msg, Service: service}
	return
}
