package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bketelsen/skynet"
	"path"
	"strings"
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
	InstanceStatsUpdateNotification
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

func NewInstanceMonitor(doozer *skynet.DoozerConnection, monitorStats bool) (im *InstanceMonitor) {
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

	if monitorStats {
		go im.monitorInstanceStats()
	}

	return
}

func (im *InstanceMonitor) mux() {

	for {
		select {
		case notification := <-im.notificationChan:

			// Update internal instance list
			switch notification.Type {
			// Stats updates don't update our internal map
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
					// Stats updates should appear as normal updates for any clients who have included Stats
					if notification.Type == InstanceStatsUpdateNotification && c.includeStats {
						notification.Type = InstanceUpdateNotification
						c.notify(notification)
					} else if notification.Type != InstanceStatsUpdateNotification {
						c.notify(notification)
					}

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
					if listener.includeStats {
						s.FetchStats(im.doozer)
					}

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

func (im *InstanceMonitor) monitorInstanceStats() {
	rev := im.doozer.GetCurrentRevision()

	watchPath := path.Join("/statistics", "**")

	for {
		ev, err := im.doozer.Wait(watchPath, rev+1)
		rev = ev.Rev

		if err != nil {
			continue
		}

		// If it's being removed no need to send notification, it's sent my monitorInstances
		if ev.IsDel() {
			continue
		} else {
			var s skynet.ServiceInfo
			var stats skynet.ServiceStatistics
			var ok bool

			servicePath := strings.Replace(ev.Path, "/statistics", "/services", 1)

			// If InstanceMonitor doesn't know about it, it was probably deleted, safe not to send notification
			if s, ok = im.instances[servicePath]; !ok {
				continue
			}

			buf := bytes.NewBuffer(ev.Body)
			err = json.Unmarshal(buf.Bytes(), &stats)

			if err != nil {
				fmt.Println("error unmarshalling service")
				continue
			}

			s.Stats = &stats

			// Let's create an update notification to send, with our new statistics
			im.notificationChan <- InstanceMonitorNotification{
				Path:       ev.Path,
				Service:    s,
				OldService: im.instances[servicePath],
				Type:       InstanceStatsUpdateNotification,
			}
		}
	}
}

func (im *InstanceMonitor) buildInstanceList(l *InstanceListener) {
	im.listChan <- l

	<-l.doneInitializing
}

func (im *InstanceMonitor) Listen(id string, q *skynet.Query, includeStats bool) (l *InstanceListener) {
	l = NewInstanceListener(im, id, q, includeStats)

	im.buildInstanceList(l)

	return
}
