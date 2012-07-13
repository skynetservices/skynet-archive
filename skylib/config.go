//Copyright (c) 2011 Brian Ketelsen

//Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package skylib

import (
	"flag"
	"fmt"
	"os"
	"time"
)

type BindAddr struct {
	IPAddress string
	Port      int
}

func (ba BindAddr) String() string {
	return fmt.Sprintf("%s:%d", ba.IPAddress, ba.Port)
}

type ServiceConfig struct {
	Log          Logger `json:"-"`
	UUID         string
	Name         string
	Version      string
	Region       string
	ServiceAddr  *BindAddr
	AdminAddr    *BindAddr
	DoozerConfig *DoozerConfig `json:"-"`
}

type ClientConfig struct {
	Log                Logger        `json:"-"`
	DoozerConfig       *DoozerConfig `json:"-"`
	ConnectionPoolSize int
	IdleTimeout        time.Duration
}

func GetServiceConfigFromFlags() *ServiceConfig {
	flagset := flag.NewFlagSet("config", flag.ContinueOnError)

	var (
		bindPort       *int    = flagset.Int("port", 9999, "tcp port to listen")
		adminPort      *int    = flagset.Int("adminport", 9998, "tcp port to listen for admin")
		bindAddr       *string = flagset.String("address", "127.0.0.1", "address to bind")
		region         *string = flagset.String("region", "unknown", "region service is located in")
		doozer         *string = flagset.String("doozer", "127.0.0.1:8046", "initial doozer instance to connect to")
		doozerBoot     *string = flagset.String("doozerboot", "127.0.0.1:8046", "initial doozer instance to connect to")
		doozerDiscover *bool   = flagset.Bool("autodiscover", true, "auto discover new doozer instances")
		uuid           *string = flagset.String("uuid", "", "UUID for this service")
	)

	flagset.Parse(os.Args[1:])

	if *uuid == "" {
		*uuid = UUID()
	}

	return &ServiceConfig{
		UUID:   *uuid,
		Region: *region,
		ServiceAddr: &BindAddr{
			IPAddress: *bindAddr,
			Port:      *bindPort,
		},
		AdminAddr: &BindAddr{
			IPAddress: *bindAddr,
			Port:      *adminPort,
		},
		DoozerConfig: &DoozerConfig{
			Uri:          *doozer,
			BootUri:      *doozerBoot,
			AutoDiscover: *doozerDiscover,
		},
	}
}
