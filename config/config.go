package config

import (
	"flag"
	"github.com/robfig/config"
	"github.com/skynetservices/skynet2/log"
	"os"
)

var defaultConfigFiles = []string{
	"/etc/skynet/skynet.conf",
	"./skynet.conf",
}

var configFile string
var uuid string
var conf *config.Config

func init() {
	flagset := flag.NewFlagSet("config", flag.ContinueOnError)
	flagset.StringVar(&configFile, "config", "", "Config File")
	flagset.StringVar(&uuid, "uuid", "", "uuid")

	args, _ := SplitFlagsetFromArgs(flagset, os.Args[1:])
	flagset.Parse(args)

	// Ensure we have a UUID
	if uuid == "" {
		uuid = NewUUID()
	}

	if configFile == "" {
		for _, f := range defaultConfigFiles {
			if _, err := os.Stat(f); err == nil {
				configFile = f
				break
			}
		}
	}

	if configFile == "" {
		log.Println(log.ERROR, "Failed to find config file")
		conf = config.NewDefault()
		return
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Println(log.ERROR, "Config file does not exist", err)
		conf = config.NewDefault()
		return
	}

	var err error
	if conf, err = config.ReadDefault(configFile); err != nil {
		conf = config.NewDefault()
		log.Fatal(err)
	}

	// Set default log level from config, this can be overriden at the service level when the service is created
	if l, err := conf.RawStringDefault("log.level"); err == nil {
		log.SetLogLevel(log.LevelFromString(l))
	}
}

func String(service, version, option string) (string, error) {
	s := getSection(service, version)

	return conf.String(s, option)
}

func Bool(service, version, option string) (bool, error) {
	s := getSection(service, version)

	return conf.Bool(s, option)
}

func Int(service, version, option string) (int, error) {
	s := getSection(service, version)

	return conf.Int(s, option)
}

func RawString(service, version, option string) (string, error) {
	s := getSection(service, version)

	return conf.RawString(s, option)
}

func RawStringDefault(option string) (string, error) {
	return conf.RawStringDefault(option)
}

func getSection(service, version string) string {
	s := service + "-" + version
	if conf.HasSection(s) {
		return s
	}

	return service
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

func UUID() string {
	return uuid
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
