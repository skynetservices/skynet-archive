package main

import (
	"github.com/petar/gozk"
	"time"
)

func main() {
	zk, session, err := zookeeper.Dial("kratos:2181", 15*time.Second)
	if err != nil {
		println("Couldn't connect: " + err.Error())
		return
	}

	defer zk.Close()

	// Wait for connection.
	event := <-session
	if event.State != zookeeper.STATE_CONNECTED {
		println("Couldn't connect")
		return
	}

	_, err = zk.Create("/counter", "0", zookeeper.EPHEMERAL, zookeeper.WorldACL(zookeeper.PERM_ALL))
	if err != nil {
		println(err.Error())
	} else {
		println("Created!")
	}
}
