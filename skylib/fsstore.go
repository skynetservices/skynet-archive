// In Skylib, for now.
package skylib

import (
	"os"
	"syscall"
	"path"
)


// FileSystem store for the Store interface.
// The Revision is just the timestamp on the lockfile.
type FsStore struct {
	basedir  string
	lockfile string
}

func NewFsStore(basedir string) *FsStore {
	lockfile := path.Join(basedir, "fsconfig.lock")
	err := os.MkdirAll(basedir, 0777)
	if err != nil {
		panic(err)
	}
	lf := CreateOrOpenEx(lockfile)
	lf.TouchClose()
	fs := &FsStore{basedir: basedir, lockfile: lockfile}
	return fs
}
// Responds with the first change made to any file matching path, a glob pattern, on or after rev.
// However, we will skip to the most recent revision.
// Also, we always set the Flag to 0.
// Todo: Make fname a glob.
func (me *FsStore) Wait(glob string, rev int64) (ev *Event, err os.Error) {
	for crev, _ := me.Rev(); crev < rev; {
		syscall.Sleep(1e8)
	}
	body, crev, _ := me.Get(glob, &rev)
	return &Event{Rev: crev, Path: glob, Body: body, Flag: 0}, nil
}
// Returns the current revision.
func (me *FsStore) Rev() (rev int64, err os.Error) {
	// No locking is necessary for Lstat.
	fi, err := os.Lstat(me.lockfile)
	if err != nil {
		panic(err)
	}
	rev = fi.Mtime_ns
	return
}
// Sets the contents of the file at path to value, as long as rev is greater than or equal to the file's revision. Returns the file's new revision.
func (me *FsStore) Set(file string, oldrev int64, body []byte) (newrev int64, err os.Error) {
	fl := OpenEx(me.lockfile)
	defer fl.TouchClose()
	fd, _ := os.Create(path.Join(me.basedir, file))
	defer fd.Close()
	fd.Write(body)
	return
}
// Del deletes the file at path if rev is greater than or equal to the file's revision.
func (me *FsStore) Del(file string, rev int64) (err os.Error) {
	fl := OpenEx(me.lockfile)
	defer fl.TouchClose()
	syscall.Unlink(path.Join(me.basedir, file))
	return
}
// Gets the contents (value) and revision (rev) of the file at path in the specified revision (rev). If rev is not provided, get uses the current revision.
func (me *FsStore) Get(file string, ignorerev *int64) (body []byte, rev int64, err os.Error) {
	if ignorerev != nil {
		panic("FsStore::Get() cannot accept a rev argument yet.")
	}
	fl := Open(me.lockfile)
	defer fl.Close()
	fd, _ := os.Open(path.Join(me.basedir, file))
	defer fd.Close()
	_, err = fd.Read(body)
	rev, _ = me.Rev()
	//body = []byte("") // Seems to be the behavior of doozer.
	return
}

var MyDC *FsStore
//var DC *doozer.Conn
