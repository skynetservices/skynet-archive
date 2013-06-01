package skynet

import (
	"testing"
)

func TestNewDoozerConnection(t *testing.T) {
	doozer := NewDoozerConnection("localhost:1234", "localhost:4321", true)

	if doozer.Config.Uri != "localhost:1234" {
		t.Error("NewDoozerConnection did not set doozer Uri")
	}

	if doozer.Config.BootUri != "localhost:4321" {
		t.Error("NewDoozerConnection did not set doozer BootUri")
	}

	if doozer.Config.AutoDiscover != true {
		t.Error("NewDoozerConnection did not set doozer AutoDiscover flag")
	}
}

func TestNewDoozerConnectionFromConfig(t *testing.T) {
	config := DoozerConfig{
		Uri:          "localhost:1234",
		BootUri:      "localhost:4321",
		AutoDiscover: true,
	}

	doozer := NewDoozerConnectionFromConfig(config)

	if doozer.Config.Uri != "localhost:1234" {
		t.Error("NewDoozerConnection did not set doozer Uri")
	}

	if doozer.Config.BootUri != "localhost:4321" {
		t.Error("NewDoozerConnection did not set doozer BootUri")
	}

	if doozer.Config.AutoDiscover != true {
		t.Error("NewDoozerConnection did not set doozer AutoDiscover flag")
	}
}
