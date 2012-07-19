package skylib

import (
	"os"
	"testing"
)

func TestGetServiceConfigFromFlags(t *testing.T) {
	os.Args = []string{"test", "--l=localhost:1234", "--region=TestRegion", "--doozer=localhost:8046", "--doozerboot=localhost:1232", "--autodiscover=true"}

	config, _ := GetServiceConfigFromFlags()

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
	os.Args = []string{"test"}

	config, _ := GetServiceConfigFromFlags()

	if config.ServiceAddr.IPAddress != "127.0.0.1" {
		t.Error("Address not set to default value")
	}

	if config.ServiceAddr.Port != 9999 {
		t.Error("Port not set to default value")
	}

	if config.Region != "unknown" {
		t.Error("Region not set to default value")
	}

	if config.DoozerConfig.Uri != "127.0.0.1:8046" {
		t.Error("DoozerUri not set to default value")
	}

	if config.DoozerConfig.BootUri != "127.0.0.1:8046" {
		t.Error("DoozerBootUri not set to default value")
	}

	if config.DoozerConfig.AutoDiscover != true {
		t.Error("DoozerAutoDiscover not set to default value")
	}
}
