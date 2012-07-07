package skylib

import (
	"log"
	"os"
	"testing"
)

func TestBasename(t *testing.T) {
	if basename("/foo") != "foo" {
		t.Error("failed to find correct basename")
	}

	if basename("/foo/bar") != "bar" {
		t.Error("failed to find correct basename")
	}

	if basename("/foo/bar/baz") != "baz" {
		t.Error("failed to find correct basename")
	}
}

func TestNewDoozerConnection(t *testing.T) {
	logger := log.New(os.Stderr, "", log.LstdFlags)

	doozer := NewDoozerConnection("localhost:1234", "localhost:4321", true, logger)

	if doozer.(*doozerConnection).Config.Uri != "localhost:1234" {
		t.Error("NewDoozerConnection did not set doozer Uri")
	}

	if doozer.(*doozerConnection).Config.BootUri != "localhost:4321" {
		t.Error("NewDoozerConnection did not set doozer BootUri")
	}

	if doozer.(*doozerConnection).Config.AutoDiscover != true {
		t.Error("NewDoozerConnection did not set doozer AutoDiscover flag")
	}

	if doozer.(*doozerConnection).Log != logger {
		t.Error("NewDoozerConnection did not set doozer log")
	}
}

func TestNewDoozerConnectionDefaultLogger(t *testing.T) {
	doozer := NewDoozerConnection("localhost:1234", "localhost:4321", true, nil)

	if doozer.(*doozerConnection).Log == nil {
		t.Error("NewDoozerConnection did not default logger")
	}
}

func TestNewDoozerConnectionFromConfig(t *testing.T) {
	logger := log.New(os.Stderr, "", log.LstdFlags)

	config := DoozerConfig{
		Uri:          "localhost:1234",
		BootUri:      "localhost:4321",
		AutoDiscover: true,
	}

	doozer := NewDoozerConnectionFromConfig(config, logger)

	if doozer.(*doozerConnection).Config.Uri != "localhost:1234" {
		t.Error("NewDoozerConnection did not set doozer Uri")
	}

	if doozer.(*doozerConnection).Config.BootUri != "localhost:4321" {
		t.Error("NewDoozerConnection did not set doozer BootUri")
	}

	if doozer.(*doozerConnection).Config.AutoDiscover != true {
		t.Error("NewDoozerConnection did not set doozer AutoDiscover flag")
	}

	if doozer.(*doozerConnection).Log != logger {
		t.Error("NewDoozerConnection did not set doozer log")
	}
}

func TestNewDoozerConnectionFromConfigDefaultLogger(t *testing.T) {
	config := DoozerConfig{
		Uri:          "localhost:1234",
		BootUri:      "localhost:4321",
		AutoDiscover: true,
	}

	doozer := NewDoozerConnectionFromConfig(config, nil)

	if doozer.(*doozerConnection).Log == nil {
		t.Error("NewDoozerConnection did not default logger")
	}
}
