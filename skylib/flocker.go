//Copyright (c) 2011 Christopher Dunn

//Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Shared and exclusive lockfiles.
// This relies on flock(), and link() (b/c of NFS quirk).
// In skylib, for now.
// Still needs a bit of testing, but should work fine.
package skylib

import (
	"log"
	"os"
	"fmt"
	"syscall"
)

const LOCK_SH = 1
const LOCK_EX = 2
const LOCK_NB = 4
const LOCK_UN = 8

type Flocker struct {
	File *os.File
}
// Create or open file in write-mode.
// Wait for exclusive lock (in case of race condition).
// Block even if current process owns the lock.
// Panic if file cannot be created.
func CreateOrOpenEx(name string) *Flocker {
	// B/c of nfs, the safe way is to create a tempfile
	// and hardlink its inode.
	tempname := fmt.Sprintf("%d.tmp", syscall.Getpid())
	file, err := os.OpenFile(tempname, syscall.O_CREAT, 0666)
	if err != nil {
		log.Panicf("Could not create tempfile '%s': %s", tempname, err)
	}
	defer syscall.Unlink(tempname)
	defer file.Close()
	errno := syscall.Link(tempname, name)
	if errno != 0 {
		//log.Panicf("Link(%s, %s) error:%v %s", tempname, name, errno, syscall.Errstr(errno))
		// We could warn, but this is not an error.
	}
	return OpenEx(name)
}

// Open file in read-write mode.
// Wait for exclusive lock.
// Block even if current process owns the lock.
// Panic if file does not exist.
func OpenEx(name string) *Flocker {
	file, err := os.OpenFile(name, syscall.O_RDWR, 0666)
	if err != nil {
		log.Panicf("Could not open file '%s' in read-write mode: %s", name, err)
	}
	if err != nil {
		log.Panic(err)
	}
	errno := syscall.Flock(file.Fd(), LOCK_EX)
	if errno != 0 {
		file.Close()
		log.Panicf("Flock LOCK_EX error:%v %s", errno, syscall.Errstr(errno))
	}
	return &Flocker{File: file}
}

// Open file in read-only mode.
// Wait for shared lock.
// Block even if current process owns the lock.
// Panic if file does not exist.
func Open(name string) *Flocker {
	file, err := os.Open(name)
	if err != nil {
		log.Panicf("Could not open file '%s': %s", name, err)
	}
	errno := syscall.Flock(file.Fd(), LOCK_SH)
	if errno != 0 {
		file.Close()
		log.Panicf("Flock LOCK_SH error:%v %s", errno, syscall.Errstr(errno))
	}
	return &Flocker{File: file}
}

// Unlock the file, if locked, and close the descriptor.
// Note: A child process may unlock a parent's file.
func (me *Flocker) Close() {
	errno := syscall.Flock(me.File.Fd(), LOCK_UN)
	if errno != 0 {
		//log.Panic(os.Errstr(errno))
		log.Panicf("Flock LOCK_UN error:%v %s", errno, syscall.Errstr(errno))
	}
	me.File.Close()
	me.File = nil
}
// Touch, then close.
// File must be opened for write.
func (me *Flocker) TouchClose() {
	//me.File.Write([]byte("x"))
	me.File.Truncate(0)
	me.Close()
}
