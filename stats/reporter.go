package stats

var reporters []Reporter

type Reporter interface {
	UpdateHostStats(stats Host)
}

func AddReporter(r Reporter) {
	reporters = append(reporters, r)
}

func UpdateHostStats(s Host) {
	for _, r := range reporters {
		r.UpdateHostStats(s)
	}
}
