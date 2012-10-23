package skynet

import (
	"bytes"
	"fmt"
	"github.com/4ad/doozer"
	"log"
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
	Getdirinfo(dir string, rev int64, off, lim int) (names []doozer.FileInfo, err error)
}

type DoozerConnection struct {
	Config     *DoozerConfig
	connection doozerconn
	Log        SemanticLogger

	connectionMutex sync.Mutex

	doozerInstances map[string]*DoozerServer
	currentInstance string

	instancesChan chan interface{}
	connChan      chan doozerconn
	dialChan      chan dialInstance

	muxing bool
}

func NewDoozerConnection(uri, boot string, discover bool,
	logger SemanticLogger) *DoozerConnection {
	return NewDoozerConnectionFromConfig(DoozerConfig{
		Uri:          uri,
		BootUri:      boot,
		AutoDiscover: discover,
	}, logger)
}

func NewDoozerConnectionFromConfig(config DoozerConfig,
	logger SemanticLogger) (d *DoozerConnection) {
	if logger == nil {
		logger = NewConsoleSemanticLogger("doozer", os.Stderr)
	}

	d = &DoozerConnection{
		Config:          &config,
		Log:             logger,
		instancesChan:   make(chan interface{}, 1),
		connChan:        make(chan doozerconn),
		dialChan:        make(chan dialInstance),
		doozerInstances: make(map[string]*DoozerServer),
	}

	return
}

type dialInstance struct {
	errch chan error
}

func (d *DoozerConnection) mux() {

	for {
		payload := &LogPayload{
			Action:     "*DoozerConnection.mux",
			Level:      DEBUG,
			ThreadName: "doozer",
			Tags:       []string{"doozer"},
		}
		select {
		case m := <-d.instancesChan:
			switch m := m.(type) {
			case DoozerDiscovered:
				// Log event
				payload.Message = fmt.Sprintf("DoozerDiscovered: %+v", m)
				err := d.Log.Log(payload)
				if err != nil {
					log.Printf("Error logging payload '%+v': %v", payload, err)
				}
				d.doozerInstances[m.DoozerServer.Key] = m.DoozerServer
			case DoozerRemoved:
				// Log event
				payload.Message = fmt.Sprintf("DoozerRemoved: %+v", m)
				err := d.Log.Log(payload)
				if err != nil {
					log.Printf("Error logging payload '%+v': %v", payload, err)
				}

				delete(d.doozerInstances, m.DoozerServer.Key)
			}
		case di := <-d.dialChan:
			di.errch <- d.dialAnInstanceMux()
		case d.connChan <- d.connection:
		}
	}
}

func (d *DoozerConnection) Connection() doozerconn {
	return <-d.connChan
}

func (d *DoozerConnection) dialAnInstance() (err error) {
	di := dialInstance{
		errch: make(chan error),
	}
	d.dialChan <- di
	err = <-di.errch
	return
}

// only call from mux()
func (d *DoozerConnection) dialAnInstanceMux() (err error) {

	if d.Config.Uri != "" && d.Config.BootUri != "" {
		err = d.dialMux(d.Config.Uri, d.Config.BootUri)
		if err == nil {
			return
		}
	}
	if d.Config.BootUri != "" {
		err = d.dialMux(d.Config.BootUri, "")
		if err == nil {
			return
		}
	}
	if d.Config.Uri != "" {
		err = d.dialMux(d.Config.Uri, "")
		if err == nil {
			return
		}
	}

	for key, dzInstance := range d.doozerInstances {
		err = d.dialMux(dzInstance.Addr, "")
		if err == nil {
			return
		}
		delete(d.doozerInstances, key)
	}
	err = fmt.Errorf("Couldn't connect to any doozer instance")
	return
}

// only call from mux()
func (d *DoozerConnection) dialMux(server string, boot string) error {
	var err error

	d.connection, err = doozer.Dial(server)
	if err != nil {
		return err
	}

	d.currentInstance = server
	//d.Log.Println("Connected to Doozer Instance: " + server)
	payload := &LogPayload{
		Action:     "*DoozerConnection.dialMux",
		Level:      DEBUG,
		ThreadName: "doozer",
		Tags:       []string{"doozer"},
	}
	connected := DoozerConnected{Addr: server}
	// Log connection
	payload.Message = fmt.Sprintf("%T: %+v", connected, connected)
	err = d.Log.Log(payload)
	if err != nil {
		return fmt.Errorf("Error logging payload '%+v': %v", payload, err)
	}

	return nil
}

func (d *DoozerConnection) recoverFromError(err interface{}) {
	payload := &LogPayload{
		Action:     "*DoozerConnection.recoverFromError",
		Level:      DEBUG,
		ThreadName: "doozer",
		Tags:       []string{"doozer"},
	}
	if err == "EOF" {
		// d.Log.Println("Lost connection to Doozer: Reconnecting...")
		connection := DoozerLostConnection{DoozerConfig: d.Config}
		payload.Message = "Lost connection to Doozer: Reconnecting... "
		payload.Message += fmt.Sprintf("%T: %+v", connection, connection)
		err := d.Log.Log(payload)
		if err != nil {
			log.Printf("Error logging payload '%+v': %v", payload, err)
		}

		dialErr := d.dialAnInstance()
		if dialErr != nil {
			payload.Message = "Couldn't reconnect to doozer"
			d.Log.Fatal(payload)
		}

	} else {
		// Don't know how to handle, go ahead and panic
		payload.Message = fmt.Sprintf("Unknown doozer error: %+v", err)
		d.Log.Fatal(payload)
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
		payload := &LogPayload{
			Action:     "*DoozerConnection.monitorCluster",
			Level:      DEBUG,
			ThreadName: "doozer",
			Tags:       []string{"doozer"},
		}
		// blocking wait call returns on a change
		ev, err := d.Wait("/ctl/cal/*", rev+1)
		if err != nil {
			payload.Message = "Error near d.Wait: " + err.Error()
			d.Log.Fatal(payload)
		}

		buf := bytes.NewBuffer(ev.Body)
		id := path.Base(ev.Path)
		rev = ev.Rev

		if ev.IsDel() || buf.String() == "" {
			if _, ok := d.doozerInstances[id]; ok {
				d.instancesChan <- DoozerRemoved{
					DoozerServer: d.doozerInstances[id],
				}
			}
		} else if buf.String() != "" {
			//if d.doozerInstances[id] == nil || d.doozerInstances[id].Key != buf.String() {
			d.instancesChan <- DoozerDiscovered{
				DoozerServer: d.getDoozerServer(buf.String()),
			}
			//}
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

func (d *DoozerConnection) Connect() {
	payload := &LogPayload{
		Action:     "*DoozerConnection.Connect",
		Level:      DEBUG,
		ThreadName: "doozer",
		Tags:       []string{"doozer"},
	}

	if d.Config == nil || (d.Config.Uri == "" && d.Config.BootUri == "") {
		payload.Message = "You must supply a doozer server and/or boot uri"
		d.Log.Fatal(payload)
	}

	if !d.muxing {
		d.muxing = true
		go d.mux()
	}

	err := d.dialAnInstance()
	if err != nil {
		payload.Message = "Failed to connect to any of the supplied "
		payload.Message += "Doozer Servers: " + err.Error()
		d.Log.Fatal(payload)
	}

	// Let's watch doozers internal config to check for new servers
	if d.Config.AutoDiscover == true {
		d.getDoozerInstances()
		go d.monitorCluster()
	}
}

func (d *DoozerConnection) getDoozerInstances() {
	rev := d.GetCurrentRevision()
	instances, _ := d.Connection().Getdir("/ctl/cal", rev, 0, -1)

	for _, i := range instances {
		rev := d.GetCurrentRevision()
		data, _, err := d.Get("/ctl/cal/"+i, rev)
		buf := bytes.NewBuffer(data)

		if err == nil && buf.String() != "" {
			d.instancesChan <- DoozerDiscovered{
				DoozerServer: d.getDoozerServer(buf.String()),
			}
		}
	}
}

func (d *DoozerConnection) GetCurrentRevision() (rev int64) {
	payload := &LogPayload{
		Action:     "*DoozerConnection.GetCurrentRevision",
		Level:      DEBUG,
		ThreadName: "doozer",
		Tags:       []string{"doozer"},
	}
	defer func() {
		if err := recover(); err != nil {
			d.recoverFromError(err)

			rev = d.GetCurrentRevision()
		}
	}()

	revision, err := d.Connection().Rev()

	if err != nil {
		payload.Message = "Error near d.Connection().Rev(): " + err.Error()
		d.Log.Fatal(payload)
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

	return d.Connection().Set(file, rev, body)
}

func (d *DoozerConnection) Del(path string, rev int64) (err error) {
	defer func() {
		if err := recover(); err != nil {
			d.recoverFromError(err)

			err = d.Del(path, rev)
		}
	}()

	return d.Connection().Del(path, rev)
}

func (d *DoozerConnection) Get(file string, rev int64) (data []byte, revision int64, err error) {
	defer func() {
		if err := recover(); err != nil {
			d.recoverFromError(err)

			data, revision, err = d.Get(file, rev)
		}
	}()

	return d.Connection().Get(file, &rev)
}

func (d *DoozerConnection) Getdir(path string, rev int64, offset int,
	limit int) (files []string, err error) {
	defer func() {
		if err := recover(); err != nil {
			d.recoverFromError(err)

			files, err = d.Getdir(path, rev, offset, limit)
		}
	}()

	return d.Connection().Getdir(path, rev, offset, limit)
}

func (d *DoozerConnection) Getdirinfo(path string, rev int64, offset int,
	limit int) (files []doozer.FileInfo, err error) {
	defer func() {
		if err := recover(); err != nil {
			d.recoverFromError(err)

			files, err = d.Getdirinfo(path, rev, offset, limit)
		}
	}()

	return d.Connection().Getdirinfo(path, rev, offset, limit)
}

func (d *DoozerConnection) Rev() (rev int64, err error) {
	defer func() {
		if err := recover(); err != nil {
			d.recoverFromError(err)

			rev, err = d.Rev()
		}
	}()

	return d.Connection().Rev()
}

func (d *DoozerConnection) Wait(glob string, rev int64) (ev doozer.Event, err error) {
	defer func() {
		if err := recover(); err != nil {
			d.recoverFromError(err)

			ev, err = d.Wait(glob, rev)
		}
	}()

	ev, err = d.Connection().Wait(glob, rev)

	return ev, err
}

func (d *DoozerConnection) Walk(rev int64, root string, v doozer.Visitor, errors chan<- error) {
	// TODO: we need to recover from failure here, but we need to make caller aware so they don't duplicate entries when we start the walk over again

	doozer.Walk(d.Connection().(*doozer.Conn), rev, root, v, errors)
}
