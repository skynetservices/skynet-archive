package skynet

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var portMutex sync.Mutex

type BindAddr struct {
	IPAddress string
	Port      int
	MaxPort   int
}

func BindAddrFromString(host string) (ba *BindAddr, err error) {
	if host == "" {
		return
	}
	split := strings.Index(host, ":")
	if split == -1 {
		err = fmt.Errorf("Must specify a port for address (got %q)", host)
		return
	}

	ba = &BindAddr{}

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

type ServiceConfig struct {
	UUID                 string
	Name                 string
	Version              string
	Region               string
	ServiceAddr          *BindAddr
	DoozerUpdateInterval time.Duration `json:"-" bson:"-"`
}

type ClientConfig struct {
	Host                      string
	Region                    string
	IdleConnectionsToInstance int
	MaxConnectionsToInstance  int
	IdleTimeout               time.Duration
	Prioritizer               func(i1, it *ServiceInfo) (i1IsBetter bool) `json:"-" bson:"-"`
}

func GetDefaultEnvVar(name, def string) (v string) {
	v = os.Getenv(name)
	if v == "" {
		v = def
	}
	return
}

func FlagsForClient(ccfg *ClientConfig, flagset *flag.FlagSet) {
	flagset.DurationVar(&ccfg.IdleTimeout, "timeout", DefaultIdleTimeout, "amount of idle time before timeout")
	flagset.IntVar(&ccfg.IdleConnectionsToInstance, "maxidle", DefaultIdleConnectionsToInstance, "maximum number of idle connections to a particular instance")
	flagset.IntVar(&ccfg.MaxConnectionsToInstance, "maxconns", DefaultMaxConnectionsToInstance, "maximum number of concurrent connections to a particular instance")
	flagset.StringVar(&ccfg.Region, "region", GetDefaultEnvVar("SKYNET_REGION", DefaultRegion), "region client is located in")
	flagset.StringVar(&ccfg.Host, "host", GetDefaultEnvVar("SKYNET_HOST", DefaultRegion), "host client is located in")
}

func GetClientConfig() (config *ClientConfig, args []string) {
	return GetClientConfigFromFlags(os.Args[1:])
}

func GetClientConfigFromFlags(argv []string) (config *ClientConfig, args []string) {
	config = &ClientConfig{}

	flagset := flag.NewFlagSet("config", flag.ContinueOnError)

	FlagsForClient(config, flagset)

	err := flagset.Parse(argv)

	args = flagset.Args()
	if err == flag.ErrHelp {
		// -help was given, pass it on to caller who
		// may decide to quit instead of continuing
		args = append(args, "-help")
	}

	return
}

func FlagsForService(scfg *ServiceConfig, flagset *flag.FlagSet) {
	flagset.StringVar(&scfg.UUID, "uuid", UUID(), "UUID for this service")
	flagset.StringVar(&scfg.Region, "region", GetDefaultEnvVar("SKYNET_REGION", DefaultRegion), "region service is located in")
	flagset.StringVar(&scfg.Version, "version", DefaultVersion, "version of service")
}

func GetServiceConfig() (config *ServiceConfig, args []string) {
	return GetServiceConfigFromFlags(os.Args[1:])
}

func ParseServiceFlags(scfg *ServiceConfig, flagset *flag.FlagSet, argv []string) (config *ServiceConfig, args []string) {

	rpcAddr := flagset.String("l", GetDefaultBindAddr(), "host:port to listen on for RPC")

	err := flagset.Parse(argv)
	args = flagset.Args()
	if err == flag.ErrHelp {
		// -help was given, pass it on to caller who
		// may decide to quit instead of continuing
		args = append(args, "-help")
	}

	rpcBA, err := BindAddrFromString(*rpcAddr)
	if err != nil {
		panic(err)
	}

	scfg.ServiceAddr = rpcBA

	return scfg, args
}

func GetServiceConfigFromFlags(argv []string) (config *ServiceConfig, args []string) {

	config = &ServiceConfig{}

	flagset := flag.NewFlagSet("config", flag.ContinueOnError)

	FlagsForService(config, flagset)

	return ParseServiceFlags(config, flagset, argv)
}

func getFlagName(f string) (name string) {
	if f[0] == '-' {
		minusCount := 1

		if f[1] == '-' {
			minusCount++
		}

		f = f[minusCount:]

		for i := 0; i < len(f); i++ {
			if f[i] == '=' || f[i] == ' ' {
				break
			}

			name += string(f[i])
		}
	}

	return
}

func SplitFlagsetFromArgs(flagset *flag.FlagSet, args []string) (flagsetArgs []string, additionalArgs []string) {
	for _, f := range args {
		if flagset.Lookup(getFlagName(f)) != nil {
			flagsetArgs = append(flagsetArgs, f)
		} else {
			additionalArgs = append(additionalArgs, f)
		}
	}

	return
}
