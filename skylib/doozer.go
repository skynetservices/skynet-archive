package skylib

import (
	"github.com/4ad/doozer"
	"log"
  "bytes"
)

type DoozerServer struct {
}

type DoozerConnection struct {
	Servers    []string
	Connection *doozer.Conn
	Log        *log.Logger
	Discover   bool

	// Internal use for discover
	doozerInstances map[string]string
}

func (d *DoozerConnection) Connect() {
	if len(d.Servers) < 1 {
		d.Log.Panic("Must supply at least 1 Doozer server to connect to")
	}

	server := d.Servers[0]
	var err error

	d.Connection, err = doozer.Dial(server)
	if err != nil {
		d.Log.Panic("Failed to connect to Doozer: " + err.Error())
	}

	// Let's watch doozers internal config to check for new servers
	if d.Discover == true {
		d.getDoozerInstances()
    go d.monitorCluster()
	}
}

func (d *DoozerConnection) GetCurrentRevision() int64 {
	revision, err := d.Connection.Rev()

	if err != nil {
		d.Log.Panic(err.Error())
	}

	return revision
}

func (d *DoozerConnection) Set(file string, rev int64, body []byte) (int64, error) {
	return d.Connection.Set(file, rev, body)
}

func (d *DoozerConnection) Del(path string, rev int64) error {
	return d.Connection.Del(path, rev)
}

func (d *DoozerConnection) Get(file string, rev *int64) ([]byte, int64, error) {
	return d.Connection.Get(file, rev)
}

func (d *DoozerConnection) Rev() (int64, error) {
	return d.Connection.Rev()
}

func (d *DoozerConnection) getDoozerInstances() {
  d.doozerInstances = make(map[string]string)

	rev := d.GetCurrentRevision()
	instances, _ := d.Connection.Getdir("/ctl/cal", rev, 0, -1)

	for _, i := range instances {
    rev := d.GetCurrentRevision()
    data, _, err := d.Get("/ctl/cal/" + i, &rev)
    buf := bytes.NewBuffer(data)

    if err == nil {
      data, _, err = d.Get("/ctl/node/" + buf.String() + "/addr", &rev)
      addrbuf := bytes.NewBuffer(data)

      if err == nil {
        // TODO: only add nodes that are writable
        d.doozerInstances[buf.String()] = addrbuf.String()
      }
    }
	}
}

func (d *DoozerConnection) monitorCluster(){
  // TODO: watch for changes to /ctl/cal and look for new nodes
  // also recover from errors, if we already have a list of nodes connect to one of them and wait there instead
}
