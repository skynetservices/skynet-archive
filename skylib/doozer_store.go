package skylib

import (
	"os"
	"log"
	"flag"
	"github.com/ha/doozer"
)


var DoozerServer *string = flag.String("doozerServer", "127.0.0.1:8046", "addr:port of doozer server")


// This is a doozer adapter to our skylib.Store interface.
// It's pretty trivial, as our API is doozer, but we
// need this because the Event structs are technically
// not the same type.
type DoozerStore struct {
	DC *doozer.Conn
}


// Constructor for DoozerStore.
// Connect based on DoozerServer cmd-line flag.
func DoozerConnect() *DoozerStore {
	var dc *doozer.Conn
	var err os.Error
	dc, err = doozer.Dial(*DoozerServer)
	if err != nil {
		log.Panic(err.String())
	}
	ds := &DoozerStore{DC: dc}
	DC = ds
	return ds
}

// Responds with the first change made to any file matching path, a glob pattern, on or after rev.
func (me *DoozerStore) Wait(glob string, rev int64) (ev *Event, err os.Error) {
	var dev doozer.Event
	dev, err = me.DC.Wait(glob, rev)
	ev = &Event{Rev: dev.Rev, Path: dev.Path, Body: dev.Body, Flag: dev.Flag}
	return
}

// Returns the current revision.
func (me *DoozerStore) Rev() (rev int64, err os.Error) {
	rev, err = me.DC.Rev()
	return
}

// Sets the contents of the file at path to value, as long as rev is greater than or equal to the file's revision. Returns the file's new revision.
func (me *DoozerStore) Set(file string, oldrev int64, body []byte) (newrev int64, err os.Error) {
	newrev, err = me.DC.Set(file, oldrev, body)
	return
}

// Del deletes the file at path if rev is greater than or equal to the file's revision.
func (me *DoozerStore) Del(file string, rev int64) (err os.Error) {
	err = me.DC.Del(file, rev)
	return
}

// Gets the contents (value) and revision (rev) of the file at path in the specified revision (rev). If rev is not provided, get uses the current revision.
func (me *DoozerStore) Get(file string, prev *int64) (body []byte, rev int64, err os.Error) {
	body, rev, err = me.DC.Get(file, prev)
	return
}
