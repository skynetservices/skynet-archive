package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/service"
	"path"
)

type InstanceMonitorNotification struct {
	Path    string
	Service service.Service
	Type    InstanceNotificationType
}

type InstanceNotificationType int

func (nt InstanceNotificationType) MarshalJSON() ([]byte, error) {
	v := "\"\""

	switch nt {
	case InstanceAddNotification:
		v = "\"InstanceAddNotification\""
	case InstanceUpdateNotification:
		v = "\"InstanceUpdateNotification\""
	case InstanceRemoveNotification:
		v = "\"InstanceRemoveNotification\""
	}

	return []byte(v), nil
}

const (
	InstanceAddNotification = iota
	InstanceUpdateNotification
	InstanceRemoveNotification
)

type InstanceMonitor struct {
	doozer           *skynet.DoozerConnection
	clients          map[string]*InstanceListener
	listChan         chan *InstanceListener
	listCloseChan    chan string
	instances        map[string]service.Service
	notificationChan chan InstanceMonitorNotification
}

func NewInstanceMonitor(doozer *skynet.DoozerConnection) (im *InstanceMonitor) {
	im = &InstanceMonitor{
		doozer:           doozer,
		clients:          make(map[string]*InstanceListener, 0),
		notificationChan: make(chan InstanceMonitorNotification, 1),
		listChan:         make(chan *InstanceListener),
		listCloseChan:    make(chan string, 1),
		instances:        make(map[string]service.Service, 0),
	}

	go im.mux()
	go im.monitorInstances()

	return
}

func (im *InstanceMonitor) mux() {
	for {
		select {
		case notification := <-im.notificationChan:

			// Update internal instance list
			switch notification.Type {
			case InstanceAddNotification, InstanceUpdateNotification:
				im.instances[notification.Path] = notification.Service
			case InstanceRemoveNotification:
				delete(im.instances, notification.Path)
			}

			for _, c := range im.clients {
				if c.query.PathMatches(notification.Path) {
					c.notify(notification)
				}
			}

		case listener := <-im.listChan:

			im.clients[listener.id] = listener

			listener.notifyEmpty()

			for path, s := range im.instances {
				if listener.query.PathMatches(path) {
					listener.notify(InstanceMonitorNotification{
						Path:    path,
						Service: s,
						Type:    InstanceAddNotification,
					})
				}
			}

			listener.doneInitializing <- true

		case lid := <-im.listCloseChan:
			c := im.clients[lid]
			close(c.NotificationChan)
			delete(im.clients, lid)

		}
	}
}

func (im *InstanceMonitor) RemoveListener(id string) {
	im.listCloseChan <- id
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
			im.notificationChan <- InstanceMonitorNotification{
				Path:    ev.Path,
				Service: im.instances[ev.Path],
				Type:    InstanceRemoveNotification,
			}
		} else {
			buf := bytes.NewBuffer(ev.Body)

			err = json.Unmarshal(buf.Bytes(), &s)

			if err != nil {
				fmt.Println("error unmarshalling service")
				continue
			}

			var notificationType InstanceNotificationType = InstanceAddNotification

			if _, ok := im.instances[ev.Path]; ok {
				notificationType = InstanceUpdateNotification
			}

			im.notificationChan <- InstanceMonitorNotification{
				Path:    ev.Path,
				Service: s,
				Type:    notificationType,
			}
		}
	}

}

func (im *InstanceMonitor) Listen(id string, q *Query) (l *InstanceListener) {
	l = NewInstanceListener(im, id, q)

	im.listChan <- l

	<-l.doneInitializing

	return
}
