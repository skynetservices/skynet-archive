package client

import (
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/service"
	"encoding/json"
	"bytes"
	"path"
  "fmt"
)

type InstanceMonitor struct {
  doozer skynet.DoozerConnection
  clients map[string]*InstanceListener
  addChan chan instance
  removeChan chan string
  listChan chan *InstanceListener
}

type instance struct {
  path string
  service service.Service
}

type InstanceListener struct {
  query *Query
  AddChan chan service.Service
  RemoveChan chan string
  monitor *InstanceMonitor
  id string

  Instances map[string]service.Service
}

func (l *InstanceListener) Close(){
  delete(l.monitor.clients, l.id)
}

func NewInstanceMonitor(doozer skynet.DoozerConnection) ( im *InstanceMonitor){
  im = &InstanceMonitor{
    doozer: doozer,
    clients: make(map[string]*InstanceListener, 0),
  }

  go im.mux()
  go im.monitorInstances()

  return
}

func (im *InstanceMonitor) mux(){
  for {
    select {
      case instance := <-im.addChan:
        fmt.Println("received add in mux")
        for _, c := range im.clients {
          if c.query.PathMatches(instance.path){
            c.AddChan <- instance.service
          }
        }

      case path := <-im.removeChan:
        fmt.Println("received remove in mux")
        for _, c := range im.clients {
          if c.query.PathMatches(path){
            c.RemoveChan <- path
          }
        }

      //case listener := <-im.listChan:
      // Set listeners list
    }
  }
}

func (im *InstanceMonitor) monitorInstances() {
	rev := im.doozer.GetCurrentRevision()

	watchPath := path.Join("/services", "**")

  /*
  // Build initial list of instances
	var ifc instanceFileCollector
	errch := make(chan error)
	doozer.Walk(rev, watchPath, &ifc, errch)

	select {
    case err := <-errch:
      fmt.Println(err)
    default:
	}

	for _, file := range ifc.files {
		buf, _, err := doozer.Get(file, rev)
		if err != nil {
      fmt.Println(err)
    default:
			continue
		}
		var s service.Service
		err = json.Unmarshal(buf, &s)
		if err != nil {
      fmt.Println("error unmarshalling service")
			continue
		}

    for _, c := range im.clients {
    }
	}
  */

  // Watch for changes
	for {
		ev, err := im.doozer.Wait(watchPath, rev+1)
		rev = ev.Rev

    var s service.Service

		if err != nil {
			continue
		}

    if ev.IsDel() {
      fmt.Println("sending remove to mux")
      im.removeChan <- ev.Path
    } else {
      buf := bytes.NewBuffer(ev.Body)

      err = json.Unmarshal(buf.Bytes(), &s)

      if err != nil {
        fmt.Println("error unmarshalling service")
        continue
      }

      fmt.Println("sending add to mux")
      im.addChan <- instance {path: ev.Path, service: s}
    }
	}

}

func (im *InstanceMonitor) Listen(id string, q *Query) (l *InstanceListener) {
  l = &InstanceListener {
    query: q,
    AddChan: make(chan service.Service),
    RemoveChan: make(chan string),
    monitor: im,
    id: id,
    Instances: make(map[string]service.Service),
  }

  im.clients[id] = l

  return
}

