package skynet

type CriteriaMatcher interface {
	Matches(s ServiceInfo) bool
}

type Criteria struct {
	Hosts      []string
	Regions    []string
	Instances  []string
	Services   []ServiceCriteria
	Registered *bool
}

type ServiceCriteria struct {
	Name    string
	Version string
}

func (sc *ServiceCriteria) String() string {
	if sc.Version == "" {
		return sc.Name
	}

	return sc.Name + ":" + sc.Version
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
	if c.Instances != nil && len(c.Instances) > 0 && !exists(c.Instances, s.UUID) {
		return false
	}

	if c.Registered != nil && s.Registered != *c.Registered {
		return false
	}

	// If no hosts were provided we assume any hosts match
	if c.Hosts != nil && len(c.Hosts) > 0 && !exists(c.Hosts, s.ServiceAddr.IPAddress) {
		return false
	}

	// If no regions were provided we assume any regions match
	if c.Regions != nil && len(c.Regions) > 0 && !exists(c.Regions, s.Region) {
		return false
	}

	// Check for service match

	if c.Regions != nil && len(c.Services) > 0 {
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

func (c *Criteria) AddInstance(uuid string) {
	if !exists(c.Instances, uuid) {
		c.Instances = append(c.Instances, uuid)
	}
}

func (c *Criteria) AddHost(host string) {
	if !exists(c.Hosts, host) {
		c.Hosts = append(c.Hosts, host)
	}
}

func (c *Criteria) AddRegion(region string) {
	if !exists(c.Regions, region) {
		c.Regions = append(c.Regions, region)
	}
}

func (c *Criteria) AddService(service ServiceCriteria) {
	for _, s := range c.Services {
		if s.Name == service.Name && s.Version == service.Version {
			return
		}
	}

	c.Services = append(c.Services, service)
}

// Returns a copy of this criteria
func (c *Criteria) Clone() *Criteria {
	criteria := new(Criteria)
	copy(c.Hosts, criteria.Hosts)
	copy(c.Regions, criteria.Regions)
	copy(c.Instances, criteria.Instances)
	copy(c.Services, criteria.Services)

	return criteria
}

func exists(haystack []string, needle string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}

	return false
}
