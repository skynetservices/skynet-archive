package stats

var reporters []Reporter

type Reporter interface {
	UpdateHostStats(host string, stats Host)
}

func AddReporter(r Reporter) {
	reporters = append(reporters, r)
}

func UpdateHostStats(host string, s Host) {
	for _, r := range reporters {
		go r.UpdateHostStats(host, s)
	}
}
