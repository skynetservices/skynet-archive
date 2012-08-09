package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/service"
	"path"
)

type InstanceMonitor struct {
	doozer     *skynet.DoozerConnection
	clients    map[string]*InstanceListener
	listChan   chan *InstanceListener
	instances  map[string]service.Service
  notificationChan chan InstanceListenerNotification
}

type instance struct {
	path    string
	service service.Service
}

type InstanceListener struct {
	query      *Query
  NotificationChan chan InstanceListenerNotification
	monitor    *InstanceMonitor
	id         string

	Instances map[string]service.Service

	doneChan chan bool
}

type InstanceListenerNotification struct {
  Path string
  Service service.Service
  Type InstanceListenerNotificationType
}

type InstanceListenerNotificationType int

const (
  InstanceListenerAddNotification = iota 
  InstanceListenerUpdateNotification
  InstanceListenerRemoveNotification
)

func (l *InstanceListener) Close() {
	delete(l.monitor.clients, l.id)
}

func NewInstanceMonitor(doozer *skynet.DoozerConnection) (im *InstanceMonitor) {
	im = &InstanceMonitor{
		doozer:     doozer,
		clients:    make(map[string]*InstanceListener, 0),
		notificationChan: make(chan InstanceListenerNotification),
		listChan:   make(chan *InstanceListener),
		instances:  make(map[string]service.Service, 0),
	}

	go im.mux()
	go im.monitorInstances()

	return
}

func (im *InstanceMonitor) mux() {
	for {
		select {
    case notification := <- im.notificationChan:

      // Update internal instance list
      switch notification.Type {
        case InstanceListenerAddNotification, InstanceListenerUpdateNotification: 
          im.instances[notification.Path] = notification.Service
        case InstanceListenerRemoveNotification:
          delete(im.instances, notification.Path)
      }

			for _, c := range im.clients {
				if c.query.PathMatches(notification.Path) {
					c.NotificationChan <- notification
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
			im.notificationChan <- InstanceListenerNotification {
        Path: ev.Path,
        Service: im.instances[ev.Path],
        Type: InstanceListenerRemoveNotification,
      }
		} else {
			buf := bytes.NewBuffer(ev.Body)

			err = json.Unmarshal(buf.Bytes(), &s)

			if err != nil {
				fmt.Println("error unmarshalling service")
				continue
			}

      var notificationType InstanceListenerNotificationType = InstanceListenerAddNotification;

      if _,ok := im.instances[ev.Path]; ok { 
        notificationType = InstanceListenerUpdateNotification
      }

			im.notificationChan <- InstanceListenerNotification {
        Path: ev.Path,
        Service: s,
        Type: notificationType,
      }
		}
	}

}

func (im *InstanceMonitor) Listen(id string, q *Query) (l *InstanceListener) {
	l = &InstanceListener{
		query:      q,
		NotificationChan: make(chan InstanceListenerNotification),
		monitor:    im,
		id:         id,
		Instances:  make(map[string]service.Service),
		doneChan:   make(chan bool),
	}

	im.listChan <- l
	<-l.doneChan

	im.clients[id] = l

	return
}
