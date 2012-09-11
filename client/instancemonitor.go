package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bketelsen/skynet"
	"path"
)

type InstanceMonitorNotification struct {
	Path       string
	Service    skynet.ServiceInfo
	OldService skynet.ServiceInfo
	Type       InstanceNotificationType
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

type instanceListRequest struct {
	q *skynet.Query
	r chan []skynet.ServiceInfo
}

type InstanceMonitor struct {
	doozer           *skynet.DoozerConnection
	clients          map[string]*InstanceListener
	ilqChan          chan instanceListRequest
	listChan         chan *InstanceListener
	listCloseChan    chan string
	instances        map[string]skynet.ServiceInfo
	notificationChan chan InstanceMonitorNotification
}

func NewInstanceMonitor(doozer *skynet.DoozerConnection) (im *InstanceMonitor) {
	im = &InstanceMonitor{
		doozer:           doozer,
		clients:          make(map[string]*InstanceListener, 0),
		notificationChan: make(chan InstanceMonitorNotification, 1),
		ilqChan:          make(chan instanceListRequest),
		listChan:         make(chan *InstanceListener),
		listCloseChan:    make(chan string, 1),
		instances:        make(map[string]skynet.ServiceInfo, 0),
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
				if notification.Service.Config == nil {
					panic("nil service config")
				}

				if c.Query.ServiceMatches(notification.Service) {
					c.notify(notification)
				} else if notification.OldService.Config != nil && c.Query.ServiceMatches(notification.OldService) {
					// Used to match, we need to send a remove notification
					notification.Type = InstanceRemoveNotification
					c.notify(notification)
				}
			}

		case ilq := <-im.ilqChan:
			ilq.r <- im.getQueryListMux(ilq.q)

		case listener := <-im.listChan:

			im.clients[listener.id] = listener

			listener.notifyEmpty()

			services := im.getQueryListMux(listener.Query)
			for _, s := range services {
				path := s.GetConfigPath()
				if listener.Query.ServiceMatches(s) {
					listener.notify(InstanceMonitorNotification{
						Path:    path,
						Service: s,
						Type:    InstanceAddNotification,
					})
				}
			}

			listener.doneInitializing <- true

		case lid := <-im.listCloseChan:
			delete(im.clients, lid)

		}
	}
}

func (im *InstanceMonitor) getQueryListMux(q *skynet.Query) (r []skynet.ServiceInfo) {
	for _, s := range im.instances {
		if q.ServiceMatches(s) {
			r = append(r, s)
		}
	}
	return
}

func (im *InstanceMonitor) GetQueryList(q *skynet.Query) (r []skynet.ServiceInfo) {
	ilq := instanceListRequest{
		q: q,
		r: make(chan []skynet.ServiceInfo, 1),
	}
	im.ilqChan <- ilq
	r = <-ilq.r
	return
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

		var s skynet.ServiceInfo
		err = json.Unmarshal(buf, &s)
		if err != nil {
			fmt.Println("error unmarshalling service")
			continue
		}

		im.instances[file] = s

		im.notificationChan <- InstanceMonitorNotification{
			Path:    file,
			Service: s,
			Type:    InstanceAddNotification,
		}
	}

	// Watch for changes

	watchPath := path.Join("/services", "**")

	for {
		ev, err := im.doozer.Wait(watchPath, rev+1)
		rev = ev.Rev

		var s skynet.ServiceInfo

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
				Path:       ev.Path,
				Service:    s,
				OldService: im.instances[ev.Path],
				Type:       notificationType,
			}
		}
	}

}

func (im *InstanceMonitor) buildInstanceList(l *InstanceListener) {
	im.listChan <- l

	<-l.doneInitializing
}

func (im *InstanceMonitor) Listen(id string, q *skynet.Query) (l *InstanceListener) {
	l = NewInstanceListener(im, id, q)

	im.buildInstanceList(l)

	return
}
