package skynet

import (
	"os"
	"testing"
)

func TestNewDoozerConnection(t *testing.T) {
	logger := NewConsoleSemanticLogger("test_logger", os.Stdout)

	doozer := NewDoozerConnection("localhost:1234", "localhost:4321", true, logger)

	if doozer.Config.Uri != "localhost:1234" {
		t.Error("NewDoozerConnection did not set doozer Uri")
	}

	if doozer.Config.BootUri != "localhost:4321" {
		t.Error("NewDoozerConnection did not set doozer BootUri")
	}

	if doozer.Config.AutoDiscover != true {
		t.Error("NewDoozerConnection did not set doozer AutoDiscover flag")
	}

	if doozer.Log != logger {
		t.Error("NewDoozerConnection did not set doozer log")
	}
}

func TestNewDoozerConnectionDefaultLogger(t *testing.T) {
	doozer := NewDoozerConnection("localhost:1234", "localhost:4321", true, nil)

	if doozer.Log == nil {
		t.Error("NewDoozerConnection did not default logger")
	}
}

func TestNewDoozerConnectionFromConfig(t *testing.T) {
	logger := NewConsoleSemanticLogger("test_logger2", os.Stdout)

	config := DoozerConfig{
		Uri:          "localhost:1234",
		BootUri:      "localhost:4321",
		AutoDiscover: true,
	}

	doozer := NewDoozerConnectionFromConfig(config, logger)

	if doozer.Config.Uri != "localhost:1234" {
		t.Error("NewDoozerConnection did not set doozer Uri")
	}

	if doozer.Config.BootUri != "localhost:4321" {
		t.Error("NewDoozerConnection did not set doozer BootUri")
	}

	if doozer.Config.AutoDiscover != true {
		t.Error("NewDoozerConnection did not set doozer AutoDiscover flag")
	}

	if doozer.Log != logger {
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

	if doozer.Log == nil {
		t.Error("NewDoozerConnection did not default logger")
	}
}
