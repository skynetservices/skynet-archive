package client

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
	query            *Query
	NotificationChan NotificationChan
	monitor          *InstanceMonitor
	id               string
}

func NewInstanceListener(im *InstanceMonitor, id string, q *Query) *InstanceListener {
	return &InstanceListener{
		query:            q,
		monitor:          im,
		id:               id,
		NotificationChan: make(NotificationChan, 1),
	}
}

func (l *InstanceListener) notify(n InstanceMonitorNotification) {
	ln := NewInstanceListenerNotification(n)

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
