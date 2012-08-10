package client

import (
	"github.com/bketelsen/skynet/service"
)

type NotificationChan chan InstanceListenerNotification

type InstanceListenerNotification map[string]InstanceMonitorNotification

func (n InstanceListenerNotification) Join(notification InstanceMonitorNotification) {
	if len(n) == 0 {
		n[notification.Path] = notification
		return
	}

	if v, ok := n[notification.Path]; ok {
		switch v.Type {
		case InstanceAddNotification:
			switch notification.Type {
			case InstanceAddNotification, InstanceRemoveNotification:
				// Current is add, new is add or remove. Replace notification
				n[notification.Path] = notification
			case InstanceUpdateNotification:
				// Current is add, new is update. Leave type as add, but replace service data
				on := n[notification.Path]
				on.Service = notification.Service
				n[notification.Path] = on
			}
		case InstanceUpdateNotification:
			// TODO:
			switch notification.Type {
			case InstanceUpdateNotification, InstanceRemoveNotification:
				// Current is update, new is update|remove. Replace notification
				n[notification.Path] = notification
			case InstanceAddNotification:
				// I'm not sure how we'd get an add, on top of an update, but let's assume we should just replace the service data
				on := n[notification.Path]
				on.Service = notification.Service
				n[notification.Path] = on
			}
		case InstanceRemoveNotification:
			// Current is a remove, doesn't matter what the new is it's safe to replace
			n[notification.Path] = notification
		}
	} else {
		n[notification.Path] = notification
	}
}

func NewInstanceListenerNotification(notification InstanceMonitorNotification) (n InstanceListenerNotification) {
	n = make(InstanceListenerNotification)
	n.Join(notification)

	return n
}

type InstanceListener struct {
	query            *Query
	NotificationChan NotificationChan
	monitor          *InstanceMonitor
	id               string

	Instances map[string]service.Service

	doneChan chan bool
}

func NewInstanceListener(im *InstanceMonitor, id string, q *Query) *InstanceListener {
	return &InstanceListener{
		query:            q,
		monitor:          im,
		id:               id,
		Instances:        make(map[string]service.Service),
		doneChan:         make(chan bool),
		NotificationChan: make(NotificationChan),
	}
}

func (l *InstanceListener) notify(n InstanceMonitorNotification) {
	for {
		select {
		case l.NotificationChan <- NewInstanceListenerNotification(n):
			return
		case on := <-l.NotificationChan:
			on.Join(n)

			l.NotificationChan <- on
		}
	}
}

func (l *InstanceListener) Close() {
	delete(l.monitor.clients, l.id)
}
