package client

import (
	"github.com/skynetservices/skynet2"
	"testing"
)

// TODO: need tests

func TestPoolClose(t *testing.T) {
	si := skynet.NewServiceInfo(nil)
	si.Registered = true
	si.ServiceAddr.IPAddress = "127.0.0.1"
	si.ServiceAddr.Port = 9000

	p := NewPool()
	p.AddInstance(si)

	p.Close()

	if len(p.servicePools) > 0 {
		t.Fatal("Close() did not close all service pools")
	}
}
