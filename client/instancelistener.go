package client

import (
	"github.com/bketelsen/skynet"
)

type NotificationChan chan InstanceListenerNotification

type InstanceListenerNotification map[string]InstanceMonitorNotification

func (n InstanceListenerNotification) Join(notification InstanceListenerNotification) InstanceListenerNotification {
	if len(n) == 0 {
		n = notification
		return n
	}

	for p, change := range notification {
		if v, ok := n[p]; ok {
			switch v.Type {
			case InstanceAddNotification:
				switch change.Type {
				case InstanceAddNotification, InstanceRemoveNotification:
					// Current is add, new is add or remove. Replace notification
					n[p] = change
				case InstanceUpdateNotification:
					// Current is add, new is update. Leave type as add, but replace service data
					on := n[p]
					on.Service = change.Service
					n[p] = on
				}
			case InstanceUpdateNotification:
				// TODO:
				switch v.Type {
				case InstanceUpdateNotification, InstanceRemoveNotification:
					// Current is update, new is update|remove. Replace notification
					n[p] = change
				case InstanceAddNotification:
					// I'm not sure how we'd get an add, on top of an update, but let's assume we should just replace the service data
					on := n[p]
					on.Service = change.Service
					n[p] = on
				}
			case InstanceRemoveNotification:
				// Current is a remove, doesn't matter what the new is it's safe to replace
				n[p] = change
			}
		} else {
			n[p] = change
		}
	}

	return n
}

func NewInstanceListenerNotification(notification InstanceMonitorNotification) (n InstanceListenerNotification) {
	n = make(InstanceListenerNotification)
	n[notification.Path] = notification

	return n
}

type InstanceListener struct {
	Query            *skynet.Query
	NotificationChan NotificationChan
	monitor          *InstanceMonitor
	id               string
	includeStats     bool
	doneInitializing chan bool
}

func NewInstanceListener(im *InstanceMonitor, id string, q *skynet.Query, includeStats bool) *InstanceListener {
	return &InstanceListener{
		Query:            q,
		monitor:          im,
		id:               id,
		includeStats:     includeStats,
		NotificationChan: make(NotificationChan, 1),
		doneInitializing: make(chan bool, 1),
	}
}

func (l *InstanceListener) notifyEmpty() {
	ln := make(InstanceListenerNotification)
	l.notifyAux(ln)
}

func (l *InstanceListener) notify(n InstanceMonitorNotification) {
	if l.includeStats && n.Service.Stats == nil {
		n.Service.FetchStats(l.monitor.doozer)
	}

	ln := NewInstanceListenerNotification(n)
	l.notifyAux(ln)
}

func (l *InstanceListener) notifyAux(ln InstanceListenerNotification) {
	for {
		select {
		case l.NotificationChan <- ln:
			return
		case on := <-l.NotificationChan:
			ln = on.Join(ln)
		}
	}
}

func (l *InstanceListener) Close() {
	l.monitor.RemoveListener(l.id)
}

func (l *InstanceListener) GetInstances() []skynet.ServiceInfo {
	return l.monitor.GetQueryList(l.Query)
}
