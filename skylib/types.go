//Copyright (c) 2011 Brian Ketelsen

//Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package skylib

import (
	"fmt"
	"container/vector"
	"os"
)


type SkynetRequest struct {
	Params map[string]interface{}
}

type SkynetResponse struct {
	Result map[string]interface{}
	Errors []string
}


/*
// RpcService is a struct that represents a remotely 
// callable function.  It is intented to be part of 
// an array or collection of RpcServices.  It contains
// a member "Provides" which is the name of the service the
// remote call provides, and a Client pointer which is a pointer
// to an RPC client connected to this service.
type RpcService struct {
	Provides string
}


func (r *RpcService) parseError(err string) {
	panic(&Error{err, r.Provides})
}
*/


// A HeartbeatRequest is the struct that is sent for ping checks.
type HeartbeatRequest struct {
	Timestamp int64
}

// HeartbeatResponse is the struct that is returned on a ping check.
type HeartbeatResponse struct {
	Timestamp int64
	Ok        bool
}

// HealthCheckRequest is the struct that is sent on a more advanced heartbeat request.
type HealthCheckRequest struct {
	Timestamp int64
}


// HealthCheckResponse is the struct that is sent back to the HealthCheckRequest-er
type HealthCheckResponse struct {
	Timestamp int64
	Load      float64
}


// A Route represents an ordered list of RPC calls that should be made for 
// a request.  Routes are versioned and named.  Names should correspond to 
// Service names- which makes me wonder if the route should be stored right there 
// in the Service struct??
type Route struct {
	Name        string
	RouteList   *vector.Vector
	Revision    int64
	LastUpdated int64
}

// The struct that is stored in the Route
// Async delineates whether it's ok to call this and not
// care about the response.
// OkToRetry delineates whether it's ok to call this service
// more than once.
type RpcCall struct {
	Service   string
	Async     bool
	OkToRetry bool
	ErrOnFail bool
}

// Parent struct for the configuration
type NetworkServers struct {
	Services []*RpcServer
}

type ServerConfig interface {
	Equal(that interface{}) bool
}

type Error struct {
	Msg     string
	Service string
}

func (e *Error) String() string { return fmt.Sprintf("Service %s had error: %s", e.Service, e.Msg) }

func NewError(msg string, service string) (err *Error) {
	err = &Error{Msg: msg, Service: service}
	return
}

// CheckError is a deferred function to turn a panic with type *Error into a plain error return.
// Other panics are unexpected and so are re-enabled.
func CheckError(error *os.Error) {
	if v := recover(); v != nil {
		if e, ok := v.(*Error); ok {
			*error = e
		} else {
			// runtime errors should crash
			panic(v)
		}
	}
}
