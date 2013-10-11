package skynet

import (
	"fmt"
	"github.com/skynetservices/skynet2/config"
	"github.com/skynetservices/skynet2/log"
	"net"
	"strconv"
	"strings"
	"sync"
)

var portMutex sync.Mutex

// ServiceStatistics contains information about its service that can
// be used to estimate load.
type ServiceStatistics struct {
	// Clients is the number of clients currently connected to this service.
	Clients int32
	// StartTime is the time when the service began running.
	StartTime string
	// LastRequest is the time when the last request was made.
	LastRequest string
}

// ServiceInfo is the publicly reported information about a particular
// service instance.
type ServiceInfo struct {
	UUID    string
	Name    string
	Version string
	Region  string

	ServiceAddr BindAddr

	// Registered indicates if the instance is currently accepting requests.
	Registered bool
}

func (si ServiceInfo) AddrString() string {
	return si.ServiceAddr.String()
}

func NewServiceInfo(name, version string) (si *ServiceInfo) {
	// TODO: we need to grab Host/Region/ServiceAddr from config
	si = &ServiceInfo{
		Name:    name,
		Version: version,
		UUID:    config.UUID(),
	}

	var host string
	var minPort, maxPort int

	if r, err := config.String(name, version, "region"); err == nil {
		si.Region = r
	} else {
		si.Region = config.DefaultRegion
	}

	if h, err := config.String(name, version, "host"); err == nil {
		host = h
	} else {
		host = config.DefaultHost
	}

	if p, err := config.Int(name, version, "service.port.min"); err == nil {
		minPort = p
	} else {
		minPort = config.DefaultMinPort
	}

	if p, err := config.Int(name, version, "service.port.max"); err == nil {
		maxPort = p
	} else {
		maxPort = config.DefaultMaxPort
	}

	log.Println(log.TRACE, host, minPort, maxPort)
	si.ServiceAddr = BindAddr{IPAddress: host, Port: minPort, MaxPort: maxPort}

	return si
}

type BindAddr struct {
	IPAddress string
	Port      int
	MaxPort   int
}

func BindAddrFromString(host string) (ba BindAddr, err error) {
	if host == "" {
		return
	}
	split := strings.Index(host, ":")
	if split == -1 {
		err = fmt.Errorf("Must specify a port for address (got %q)", host)
		return
	}

	ba = BindAddr{}

	ba.IPAddress = host[:split]
	if ba.IPAddress == "" {
		ba.IPAddress = "0.0.0.0"
	}

	portstr := host[split+1:]
	if ba.Port, err = strconv.Atoi(portstr); err == nil {
		return
	}

	var rindex int
	if rindex = strings.Index(portstr, "-"); rindex == -1 {
		err = fmt.Errorf("Couldn't process port for %q: %v", host, err)
		return
	}

	maxPortStr := portstr[rindex+1:]
	portstr = portstr[:rindex]

	if ba.Port, err = strconv.Atoi(portstr); err != nil {
		err = fmt.Errorf("Couldn't process port for %q: %v", host, err)
		return
	}
	if ba.MaxPort, err = strconv.Atoi(maxPortStr); err != nil {
		err = fmt.Errorf("Couldn't process port for %q: %v", host, err)
		return
	}

	return
}

func (ba *BindAddr) String() string {
	if ba == nil {
		return ""
	}
	return fmt.Sprintf("%s:%d", ba.IPAddress, ba.Port)
}

func (ba *BindAddr) Listen() (listener *net.TCPListener, err error) {
	// Ensure Admin, and RPC don't fight over the same port
	portMutex.Lock()
	defer portMutex.Unlock()

	for {
		var laddr *net.TCPAddr
		laddr, err = net.ResolveTCPAddr("tcp", ba.String())
		if err != nil {
			panic(err)
		}
		listener, err = net.ListenTCP("tcp", laddr)
		if err == nil {
			return
		}
		if ba.Port < ba.MaxPort {
			ba.Port++
		} else {
			return
		}

	}
	return
}
