package client

import (
	"github.com/bketelsen/skynet/service"
)

type InstanceListener struct {
	query            *Query
	NotificationChan chan InstanceListenerNotification
	monitor          *InstanceMonitor
	id               string
	callback         InstanceListenerCallback

	Instances map[string]service.Service

	doneChan chan bool
}

type InstanceListenerCallback func(n InstanceListenerNotification)

type InstanceListenerNotification struct {
	Path    string
	Service service.Service
	Type    InstanceListenerNotificationType
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
