package skynet

type CriteriaMatcher interface {
	Matches(s ServiceInfo) bool
}

type Criteria struct {
	Hosts      []string
	Regions    []string
	Services   []ServiceCriteria
	Registered *bool
}

type ServiceCriteria struct {
	Name    string
	Version string
}

func (sc *ServiceCriteria) Matches(name, version string) bool {
	if sc.Name != "" && sc.Name != name {
		return false
	}

	if sc.Version != "" && sc.Version != version {
		return false
	}

	return true
}

func (c *Criteria) Matches(s ServiceInfo) bool {
	if c.Registered != nil && s.Registered != *c.Registered {
		return false
	}

	// If no hosts were provided we assume any hosts match
	if len(c.Hosts) > 0 && !exists(c.Hosts, s.ServiceAddr.IPAddress) {
		return false
	}

	// If no regions were provided we assume any regions match
	if len(c.Regions) > 0 && !exists(c.Regions, s.Region) {
		return false
	}

	// Check for service match

	if len(c.Services) > 0 {
		match := false
		for _, sc := range c.Services {
			if sc.Matches(s.Name, s.Version) {
				match = true
				break
			}
		}

		if !match {
			return false
		}
	}

	return true
}

func exists(haystack []string, needle string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}

	return false
}
