package skylib

import "os"

// This is returned by Store::Wait().
// Just like doozer.Event.
type Event struct {
	Rev  int64  // revision of the change
	Path string // actual file (in case of wildcards)
	Body []byte // contents at Rev
	Flag int32  // 4=>changed-or-created; 8=>deleted
}

// These signatures are copied from doozerd.
// https://github.com/ha/doozerd/blob/master/doc/proto.md
type Store interface {
	Wait(glob string, rev int64) (ev *Event, err os.Error)
	Rev() (int64, os.Error)
	//Close()
	//Access(token string) os.Error
	Set(file string, oldrev int64, body []byte) (newrev int64, err os.Error)
	Del(file string, rev int64) os.Error
	//Nop() os.Error
	Get(file string, rev *int64) ([]byte, int64, os.Error)
	//Getdir(dir string, rev int64, off, lim int) (names []string, err os.Error)
	//Stat(path string, storeRev *int64) (len int, fileRev int64, err os.Error)
}
