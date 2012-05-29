package skylib

import (
	"bytes"
	"github.com/4ad/doozer"
	"log"
	"sync"
)

type DoozerServer struct {
	Key  string
	Id   int
	Addr string
}

type DoozerConnection struct {
  Config     *DoozerConfig
	Connection *doozer.Conn
	Log        *log.Logger

	connectionMutex sync.Mutex

	// Internal use for discover
	doozerInstances map[string]*DoozerServer
	currentInstance string
}

type DoozerConfig struct {
	Uri          string
	BootUri      string
	AutoDiscover bool
}

func NewDoozerConnection(uri string, boot string, discover bool) (*DoozerConnection){
  return &DoozerConnection {
    Config:  &DoozerConfig {
      Uri: uri,
      BootUri: boot,
      AutoDiscover: discover,
    },
  }
}

func (d *DoozerConnection) Connect() {
	if d.Config == nil || (d.Config.Uri == "" && d.Config.BootUri == ""){
		d.Log.Panic("You must supply a doozer server or/and boot uri")
	}

	var success = false
	var err error = nil

  if d.Config.Uri != "" && d.Config.BootUri != "" {
    success, err = d.dial(d.Config.Uri, d.Config.BootUri)
  } else if d.Config.BootUri != "" {
    success, err = d.dial(d.Config.BootUri, "")
  } else {
    success, err = d.dial(d.Config.Uri, "")
  }

	if success == false {
		d.Log.Panic("Failed to connect to any of the supplied Doozer Servers: " + err.Error())
	}

	// Let's watch doozers internal config to check for new servers
	if d.Config.AutoDiscover == true {
		d.getDoozerInstances()
		go d.monitorCluster()
	}
}

func (d *DoozerConnection) dial(server string, boot string) (bool, error) {
	var err error

	d.Connection, err = doozer.Dial(server)
	if err != nil {
		return false, err
	}

	d.currentInstance = server
	d.Log.Println("Connected to Doozer Instance: " + server)

	return true, nil
}

func (d *DoozerConnection) GetCurrentRevision() (rev int64) {
	defer func() {
		if err := recover(); err != nil {
			d.recoverFromError(err)

			rev = d.GetCurrentRevision()
		}
	}()

	revision, err := d.Connection.Rev()

	if err != nil {
		d.Log.Panic(err.Error())
	}

	return revision
}

func (d *DoozerConnection) Set(file string, rev int64, body []byte) (newRev int64, err error) {
	defer func() {
		if err := recover(); err != nil {
			d.recoverFromError(err)

			newRev, err = d.Set(file, rev, body)
		}
	}()

	return d.Connection.Set(file, rev, body)
}

func (d *DoozerConnection) Del(path string, rev int64) (err error) {
	defer func() {
		if err := recover(); err != nil {
			d.recoverFromError(err)

			err = d.Del(path, rev)
		}
	}()

	return d.Connection.Del(path, rev)
}

func (d *DoozerConnection) Get(file string, rev *int64) (data []byte, revision int64, err error) {
	defer func() {
		if err := recover(); err != nil {
			d.recoverFromError(err)

			data, revision, err = d.Get(file, rev)
		}
	}()

	return d.Connection.Get(file, rev)
}

func (d *DoozerConnection) Rev() (rev int64, err error) {
	defer func() {
		if err := recover(); err != nil {
			d.recoverFromError(err)

			rev, err = d.Rev()
		}
	}()


	return d.Connection.Rev()
}

func (d *DoozerConnection) getDoozerInstances() {
	d.doozerInstances = make(map[string]*DoozerServer)

	rev := d.GetCurrentRevision()
	instances, _ := d.Connection.Getdir("/ctl/cal", rev, 0, -1)

	for _, i := range instances {
		rev := d.GetCurrentRevision()
		data, _, err := d.Get("/ctl/cal/"+i, &rev)
		buf := bytes.NewBuffer(data)

		if err == nil {
			d.doozerInstances[i] = d.getDoozerServer(buf.String())
		}
	}
}

func (d *DoozerConnection) recoverFromError(err interface{}) {
	if err == "EOF" {
		d.Log.Println("Lost connection to Doozer: Reconnecting...")
		d.connectionMutex.Lock()
		defer d.connectionMutex.Unlock()

		// if they enabled Auto Discovery let's try to get a connection from one of the instances we know about
		if len(d.doozerInstances) > 0 && d.Config.AutoDiscover == true {
			for _, server := range d.doozerInstances {
				success, _ := d.dial(server.Addr, "")

				if success == true {
					return
				}
			}
		}

		// If we made it here we didn't find a server
		d.Log.Panic("Unable to find a Doozer instance to connect to")

	} else {
		// Don't know how to handle, go ahead and panic
		d.Log.Panic(err)
	}
}

func (d *DoozerConnection) monitorCluster() {
	defer func() {
		if err := recover(); err != nil {
			d.recoverFromError(err)

			d.monitorCluster()
		}
	}()

	for {
		// blocking wait call returns on a change
		ev, err := d.Connection.Wait("/ctl/cal/*", d.GetCurrentRevision())
		if err != nil {
			d.Log.Panic(err.Error())
		}

		buf := bytes.NewBuffer(ev.Body)
		id := basename(ev.Path)

		if buf.String() == "" && d.doozerInstances[id] != nil {
			// Server is down, remove from list
			d.Log.Println("Doozer instance no longer available, removing from available list")
			delete(d.doozerInstances, id)

		} else if buf.String() != "" {
			// Server changed, check to make sure it's different first
			if d.doozerInstances[id] == nil || d.doozerInstances[id].Key != buf.String() {
				d.Log.Println("New Doozer instance detected, adding to available list")

				d.doozerInstances[id] = d.getDoozerServer(buf.String())
			}
		}
	}
}

func (d *DoozerConnection) getDoozerServer(key string) *DoozerServer {
	rev := d.GetCurrentRevision()
	data, _, err := d.Get("/ctl/node/"+key+"/addr", &rev)
	buf := bytes.NewBuffer(data)

	if err == nil {
		return &DoozerServer{
			Addr: buf.String(),
			Key:  key,
		}
	}

	return nil
}

func basename(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}
