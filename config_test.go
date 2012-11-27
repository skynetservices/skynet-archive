package skynet

import (
	"fmt"
	"os"
	"testing"
)

func clearEnv() {
	os.Setenv("SKYNET_BIND_IP", "")
	os.Setenv("SKYNET_MIN_PORT", "")
	os.Setenv("SKYNET_MAX_PORT", "")
	os.Setenv("SKYNET_DZHOST", "")
	os.Setenv("SKYNET_DZNSHOST", "")
	os.Setenv("SKYNET_DZDISCOVER", "")
	os.Setenv("SKYNET_MGOSERVER", "")
	os.Setenv("SKYNET_MGODB", "")
	os.Setenv("SKYNET_REGION", "")
	os.Setenv("SKYNET_VERSION", "")
}

func TestGetServiceConfigFromFlags(t *testing.T) {
	clearEnv()

	os.Args = []string{"test", "--l=localhost:1234", "--region=TestRegion",
		"--doozer=localhost:8046", "--doozerboot=localhost:1232",
		"--autodiscover=true",
	}

	config, _ := GetServiceConfigFromFlags(os.Args[1:])

	if config.ServiceAddr.IPAddress != "localhost" {
		t.Error("Address not set through flag")
	}

	if config.ServiceAddr.Port != 1234 {
		t.Error("Port not set through flag")
	}

	if config.Region != "TestRegion" {
		t.Error("Region not set through flag")
	}

	if config.DoozerConfig.Uri != "localhost:8046" {
		t.Error("DoozerUri not set through flag")
	}

	if config.DoozerConfig.BootUri != "localhost:1232" {
		t.Error("DoozerBootUri not set through flag")
	}

	if config.DoozerConfig.AutoDiscover != true {
		t.Error("DoozerAutoDiscover not set through flag")
	}
}

func TestGetServiceConfigFromFlagsDefaults(t *testing.T) {
	clearEnv()

	os.Args = []string{"test"}

	config, _ := GetServiceConfigFromFlags(os.Args[1:])

	if config.ServiceAddr.IPAddress != "127.0.0.1" {
		t.Error("Address not set to default value")
	}

	if config.ServiceAddr.Port != 9000 {
		t.Error("Port not set to default value")
	}

	if config.Region != "unknown" {
		t.Error("Region not set to default value")
	}

	if config.DoozerConfig.Uri != "127.0.0.1:8046" {
		fmt.Println(config.DoozerConfig.Uri)
		t.Error("DoozerUri not set to default value")
	}

	if config.DoozerConfig.BootUri != "127.0.0.1:8046" {
		t.Error("DoozerBootUri not set to default value")
	}

	if config.DoozerConfig.AutoDiscover != true {
		t.Error("DoozerAutoDiscover not set to default value")
	}
}
