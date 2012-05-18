//Copyright (c) 2011 Brian Ketelsen

//Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package skylib

import (
	"log"
	"flag"
  "fmt"
)

type BindAddr struct {
	IPAddress string
	Port      int
}

type Config struct {
	Log                   *log.Logger
	Name                  string
	Version               string
	Region                string
  ServiceAddr           *BindAddr
  AdminAddr             *BindAddr

	ConfigServers         []string  `json:"-"`
	ConfigServerDiscovery bool      `json:"-"`
}

type DoozerConfig []string

func (dc *DoozerConfig) Set(s string) error {
	*dc = append(*dc, s)
	return nil
}

func (dc *DoozerConfig) String() string {
	return fmt.Sprint(*dc)
}



func GetConfigFromFlags() (*Config){
  var (
    bindPort    *int    = flag.Int("port", 9999, "tcp port to listen")
    bindAddr    *string = flag.String("address", "127.0.0.1", "address to bind")
    region      *string = flag.String("region", "unknown", "region service is located in")
    doozerDiscover *bool = flag.Bool("autodiscover", true, "auto discover new doozer instances")
    doozerAddrs         = DoozerConfig{}
  )

	flag.Var(&doozerAddrs, "doozer", "addr:port of doozer server") // trick to supply multiple -doozer flags

	dzServers := make([]string, 0)
	for _, dz := range doozerAddrs {
		log.Println(dz)
		dzServers = append(dzServers, dz)
	}

	flag.Parse()
  return &Config{
      Region: *region,
      ServiceAddr:           &BindAddr {
        IPAddress:             *bindAddr,
        Port:                  *bindPort,
      },
      ConfigServers:         dzServers,
      ConfigServerDiscovery: *doozerDiscover,
    }
}
