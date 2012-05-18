package skylib

import (
	"bytes"
	"github.com/4ad/doozer"
	"log"
  "sync"
)

type DoozerServer struct {
	Key      string
	Id       int
	Addr     string
}

type DoozerConnection struct {
	Servers    []string
	Connection *doozer.Conn
	Log        *log.Logger
	Discover   bool

  connectionMutex sync.Mutex

	// Internal use for discover
	doozerInstances map[string]*DoozerServer
  currentInstance string
}

// TODO: Need to support booturi
func (d *DoozerConnection) Connect() {
	if len(d.Servers) < 1 {
		d.Log.Panic("Must supply at least 1 Doozer server to connect to")
	}

  var success = false
  var err error = nil

  for _, server := range d.Servers {
    success, err = d.dial(server)

    if success == true {
      break
    }
  }

  if success == false {
    d.Log.Panic("Failed to connect to any of the supplied Doozer Servers: " + err.Error())
  }

	// Let's watch doozers internal config to check for new servers
	if d.Discover == true {
		d.getDoozerInstances()
		go d.monitorCluster()
	}
}

func (d *DoozerConnection) dial(server string)  (bool, error) {
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

func (d *DoozerConnection) recoverFromError(err interface{}){
  if err == "EOF" {
    d.Log.Println("Lost connection to Doozer: Reconnecting...")
    d.connectionMutex.Lock()
    defer d.connectionMutex.Unlock()


    // Let's try to connect to the servers supplied in config
    for _, server := range d.Servers {
      success, _ := d.dial(server)

      if success == true {
        return
      }
    }

    // If we didn't connect to one of the initially supplied servers and they enabled Auto Discovery
    // Let's try to get a connection from one of the instances we know about
    if len(d.doozerInstances) > 0 && d.Discover == true {
      for _, server := range d.doozerInstances {
        success, _ := d.dial(server.Addr)

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

func (d *DoozerConnection) getDoozerServer(key string) (*DoozerServer){
  rev := d.GetCurrentRevision()
  data, _, err := d.Get("/ctl/node/" + key + "/addr", &rev)
  buf := bytes.NewBuffer(data)

  if err == nil {
    return &DoozerServer {
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
