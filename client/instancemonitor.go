package client

import (
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/service"
	"encoding/json"
	"bytes"
  "fmt"
  "path"
)

type InstanceMonitor struct {
  doozer skynet.DoozerConnection
  clients map[string]*InstanceListener
  addChan chan instance
  removeChan chan string
  listChan chan *InstanceListener
  instances map[string]service.Service
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

  doneChan chan bool
}

func (l *InstanceListener) Close(){
  delete(l.monitor.clients, l.id)
}

func NewInstanceMonitor(doozer skynet.DoozerConnection) ( im *InstanceMonitor){
  im = &InstanceMonitor{
    doozer: doozer,
    clients: make(map[string]*InstanceListener, 0),
    addChan: make(chan instance),
    removeChan: make(chan string),
    listChan: make(chan *InstanceListener),
    instances: make(map[string]service.Service, 0),
  }

  go im.mux()
  go im.monitorInstances()

  return
}

func (im *InstanceMonitor) mux(){
  for {
    select {
      case instance := <-im.addChan:
        for _, c := range im.clients {

          im.instances[instance.path] = instance.service

          if c.query.PathMatches(instance.path){
            c.AddChan <- instance.service
          }
        }

      case path := <-im.removeChan:
        delete(im.instances, path)

        for _, c := range im.clients {
          if c.query.PathMatches(path){
            c.RemoveChan <- path
          }
        }

      case listener := <-im.listChan:
        for path, s := range im.instances {
          if listener.query.PathMatches(path) {
            listener.Instances[path] = s
          }
        }

        listener.doneChan <- true
    }
  }
}

func (im *InstanceMonitor) monitorInstances() {
	rev := im.doozer.GetCurrentRevision()

  // Build initial list of instances
	var ifc instanceFileCollector
	errch := make(chan error)
	im.doozer.Walk(rev, "/services", &ifc, errch)

	select {
    case err := <-errch:
      fmt.Println(err)
    default:
	}

	for _, file := range ifc.files {
		buf, _, err := im.doozer.Get(file, rev)
		if err != nil {
      fmt.Println(err)
      continue
    }

		var s service.Service
		err = json.Unmarshal(buf, &s)
		if err != nil {
      fmt.Println("error unmarshalling service")
			continue
		}

    im.instances[file] = s
	}

  // Watch for changes

	watchPath := path.Join("/services", "**")

	for {
		ev, err := im.doozer.Wait(watchPath, rev+1)
		rev = ev.Rev

    var s service.Service

		if err != nil {
			continue
		}

    if ev.IsDel() {
      im.removeChan <- ev.Path
    } else {
      buf := bytes.NewBuffer(ev.Body)

      err = json.Unmarshal(buf.Bytes(), &s)

      if err != nil {
        fmt.Println("error unmarshalling service")
        continue
      }


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
    doneChan: make(chan bool),
  }

  im.listChan <- l
  <-l.doneChan

  im.clients[id] = l

  return
}

