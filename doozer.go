package skynet

import (
	"bytes"
	"github.com/4ad/doozer"
	"os"
	"path"
	"sync"
)

type DoozerServer struct {
	Key  string
	Id   int
	Addr string
}

type DoozerConfig struct {
	Uri          string
	BootUri      string
	AutoDiscover bool
}

// Used as interface to doozer.Conn so that we can stub for tests
type doozerconn interface {
	Set(file string, rev int64, body []byte) (newRev int64, err error)
	Del(path string, rev int64) (err error)
	Get(file string, rev *int64) (data []byte, revision int64, err error)
	Wait(glob string, rev int64) (ev doozer.Event, err error)
	Walk(string, int64, int, int) ([]doozer.Event, error)
	Rev() (rev int64, err error)
	Getdir(dir string, rev int64, off, lim int) (names []string, err error)
}

type DoozerConnection struct {
	Config     *DoozerConfig
	Connection doozerconn
	Log        Logger

	connectionMutex sync.Mutex

	doozerInstances map[string]*DoozerServer
	currentInstance string
}

func NewDoozerConnection(uri string, boot string, discover bool, logger Logger) *DoozerConnection {
	if logger == nil {
		logger = NewConsoleLogger(os.Stderr)
	}

	return &DoozerConnection{
		Config: &DoozerConfig{
			Uri:          uri,
			BootUri:      boot,
			AutoDiscover: discover,
		},

		Log: logger,
	}
}

func NewDoozerConnectionFromConfig(config DoozerConfig, logger Logger) *DoozerConnection {
	if logger == nil {
		logger = NewConsoleLogger(os.Stderr)
	}

	return &DoozerConnection{
		Config: &config,
		Log:    logger,
	}
}

func (d *DoozerConnection) Connect() {
	if d.Config == nil || (d.Config.Uri == "" && d.Config.BootUri == "") {
		d.Log.Panic("You must supply a doozer server and/or boot uri")
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

func (d *DoozerConnection) Get(file string, rev int64) (data []byte, revision int64, err error) {
	defer func() {
		if err := recover(); err != nil {
			d.recoverFromError(err)

			data, revision, err = d.Get(file, rev)
		}
	}()

	return d.Connection.Get(file, &rev)
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

func (d *DoozerConnection) Wait(glob string, rev int64) (ev doozer.Event, err error) {
	defer func() {
		if err := recover(); err != nil {
			d.recoverFromError(err)

			ev, err = d.Wait(glob, rev)
		}
	}()

	ev, err = d.Connection.Wait(glob, rev)

	return ev, err
}

func (d *DoozerConnection) Walk(rev int64, root string, v doozer.Visitor, errors chan<- error) {
	// TODO: we need to recover from failure here, but we need to make caller aware so they don't duplicate entries when we start the walk over again

	doozer.Walk(d.Connection.(*doozer.Conn), rev, root, v, errors)
}

func (d *DoozerConnection) dial(server string, boot string) (bool, error) {
	var err error

	d.Connection, err = doozer.Dial(server)
	if err != nil {
		return false, err
	}

	d.currentInstance = server
	//d.Log.Println("Connected to Doozer Instance: " + server)
	d.Log.Item(DoozerConnected{
		Addr: server,
	})

	return true, nil
}

func (d *DoozerConnection) getDoozerInstances() {
	d.doozerInstances = make(map[string]*DoozerServer)

	rev := d.GetCurrentRevision()
	instances, _ := d.Connection.Getdir("/ctl/cal", rev, 0, -1)

	for _, i := range instances {
		rev := d.GetCurrentRevision()
		data, _, err := d.Get("/ctl/cal/"+i, rev)
		buf := bytes.NewBuffer(data)

		if err == nil {
			d.doozerInstances[i] = d.getDoozerServer(buf.String())
		}
	}
}

func (d *DoozerConnection) recoverFromError(err interface{}) {
	if err == "EOF" {
		// d.Log.Println("Lost connection to Doozer: Reconnecting...")
		d.Log.Item(DoozerLostConnection{
			DoozerConfig: d.Config,
		})
		d.connectionMutex.Lock()
		defer d.connectionMutex.Unlock()

		// if they enabled Auto Discovery let's try to get a connection from one of the instances we know about
		if len(d.doozerInstances) > 0 && d.Config.AutoDiscover == true {
			for key, server := range d.doozerInstances {
				success, _ := d.dial(server.Addr, "")

				if success == true {
					return
				} else {
					// Remove failed doozer instance from map
					delete(d.doozerInstances, key)

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

// TODO: Need to track last known revision, so when we are monitor for changes to the doozer cluster
// we can replay changes that took place while we were looking for a new connection instead of using the latest GetCurrentRevision()
func (d *DoozerConnection) monitorCluster() {
	defer func() {
		if err := recover(); err != nil {
			d.recoverFromError(err)

			d.monitorCluster()
		}
	}()

	rev := d.GetCurrentRevision()

	for {
		// blocking wait call returns on a change
		ev, err := d.Wait("/ctl/cal/*", rev+1)
		if err != nil {
			d.Log.Panic(err.Error())
		}

		buf := bytes.NewBuffer(ev.Body)
		id := path.Base(ev.Path)
		rev = ev.Rev

		if buf.String() == "" && d.doozerInstances[id] != nil {
			// Server is down, remove from list
			//d.Log.Println("Doozer instance no longer available, removing from available list")
			d.Log.Item(DoozerRemoved{
				DoozerServer: d.doozerInstances[id],
			})
			delete(d.doozerInstances, id)

		} else if buf.String() != "" {
			// Server changed, check to make sure it's different first
			if d.doozerInstances[id] == nil || d.doozerInstances[id].Key != buf.String() {
				//d.Log.Println("New Doozer instance detected, adding to available list")
				d.doozerInstances[id] = d.getDoozerServer(buf.String())
				d.Log.Item(DoozerDiscovered{
					DoozerServer: d.doozerInstances[id],
				})
			}
		}
	}
}

func (d *DoozerConnection) getDoozerServer(key string) *DoozerServer {
	rev := d.GetCurrentRevision()
	data, _, err := d.Get("/ctl/node/"+key+"/addr", rev)
	buf := bytes.NewBuffer(data)

	if err == nil {
		return &DoozerServer{
			Addr: buf.String(),
			Key:  key,
		}
	}

	return nil
}
