package client

import (
	"github.com/bketelsen/skynet"
	"strconv"
	"testing"
	"time"
)

func stubServiceInfo() (si *skynet.ServiceInfo) {
	si = &skynet.ServiceInfo{
		Config: &skynet.ServiceConfig{
			Name:        "LoadBalancer",
			Version:     "1",
			ServiceAddr: &skynet.BindAddr{},
		},
		Stats: &skynet.ServiceStatistics{},
	}

	return
}

func stubClient() (c *Client) {
	c = &Client{
		Config: &skynet.ClientConfig{},
	}

	return
}

func TestDefaultComparatorHostFirst(t *testing.T) {
	c := stubClient()
	c.Config.Region = "A"
	c.Config.Host = "127.0.0.1"

	s1 := stubServiceInfo()
	s1.Config.Region = "A"
	s1.Config.ServiceAddr.IPAddress = "127.0.0.1"

	s2 := stubServiceInfo()
	s2.Config.Region = "A"
	s2.Config.ServiceAddr.IPAddress = "192.168.1.1"

	// We should choose instances on the client's host over region
	if !defaultComparator(c, s1, s2) || defaultComparator(c, s2, s1) {
		t.Error("Failed to select instance on same host")
	}

	s2.Config.ServiceAddr.IPAddress = "127.0.0.1"
	s1.Stats.LastRequest = "2012-11-14T15:04:05Z-0700"
	s2.Stats.LastRequest = "2012-11-14T15:04:10Z-0700"

	// Tie breaker LastRequest
	if !defaultComparator(c, s1, s2) {
		t.Error("Failed to select instance with older LastRequest")
	}
}

func TestDefaultComparatorRegionFirst(t *testing.T) {
	c := stubClient()
	c.Config.Region = "A"

	s1 := stubServiceInfo()
	s1.Config.Region = "A"

	s2 := stubServiceInfo()
	s2.Config.Region = "B"

	// We should choose instances in the client's region over external
	if !defaultComparator(c, s1, s2) || defaultComparator(c, s2, s1) {
		t.Error("Failed to select instance in same region")
	}

	s2.Config.Region = "A"
	s1.Stats.LastRequest = "2012-11-14T15:04:05Z-0700"
	s2.Stats.LastRequest = "2012-11-14T15:04:10Z-0700"

	// Tie breaker LastRequest
	if !defaultComparator(c, s1, s2) {
		t.Error("Failed to select instance with older LastRequest")
	}
}

func TestDefaultComparatorCriticalClients(t *testing.T) {
	c := stubClient()
	c.Config.Region = "A"
	c.Config.Host = "127.0.0.1"

	s1 := stubServiceInfo()
	s1.Config.Region = "A"
	s1.Config.CriticalClientCount = 5
	s1.Stats.Clients = 2
	s1.Config.ServiceAddr.IPAddress = "192.168.1.1"

	s2 := stubServiceInfo()
	s2.Config.Region = "A"
	s2.Config.CriticalClientCount = 5
	s2.Stats.Clients = 6
	s2.Config.ServiceAddr.IPAddress = "192.168.1.2"

	if !defaultComparator(c, s1, s2) || defaultComparator(c, s2, s1) {
		t.Error("Failed to select instance without critical clients")
	}

	s1.Stats.LastRequest = "2012-11-14T15:04:05Z-0700"
	s2.Stats.LastRequest = "2012-11-14T15:04:10Z-0700"

	// Tie breaker LastRequest
	if !defaultComparator(c, s1, s2) {
		t.Error("Failed to select instance with older LastRequest")
	}

	// Edge case, try both in different region
	s1.Config.Region = "B"
	s2.Config.Region = "B"

	if !defaultComparator(c, s1, s2) || defaultComparator(c, s2, s1) {
		t.Error("Failed to select instance without critical clients")
	}
}

func TestDefaultComparatorCriticalClientsLeaveRegion(t *testing.T) {
	c := stubClient()
	c.Config.Region = "A"
	c.Config.Host = "127.0.0.1"

	s1 := stubServiceInfo()
	s1.Config.Region = "B"
	s1.Config.CriticalClientCount = 5
	s1.Stats.Clients = 2
	s1.Config.ServiceAddr.IPAddress = "192.168.1.1"

	s2 := stubServiceInfo()
	s2.Config.Region = "A"
	s2.Config.CriticalClientCount = 5
	s2.Stats.Clients = 6
	s2.Config.ServiceAddr.IPAddress = "192.168.1.2"

	// We should choose instances in the client's region over external
	if !defaultComparator(c, s1, s2) || defaultComparator(c, s2, s1) {
		t.Error("Failed to select instance without critical clients")
	}

	s1.Stats.LastRequest = "2012-11-14T15:04:05Z-0700"
	s2.Stats.LastRequest = "2012-11-14T15:04:10Z-0700"

	// Tie breaker LastRequest
	if !defaultComparator(c, s1, s2) {
		t.Error("Failed to select instance with older LastRequest")
	}
}

func TestDefaultComparatorCriticalClientsZeroValue(t *testing.T) {
	c := stubClient()
	c.Config.Region = "A"
	c.Config.Host = "127.0.0.1"

	s1 := stubServiceInfo()
	s1.Config.Region = "A"
	s1.Stats.Clients = 2
	s1.Config.ServiceAddr.IPAddress = "192.168.1.1"

	s2 := stubServiceInfo()
	s2.Config.Region = "B"
	s2.Stats.Clients = 0
	s2.Config.ServiceAddr.IPAddress = "192.168.1.2"

	if !defaultComparator(c, s1, s2) || defaultComparator(c, s2, s1) {
		t.Error("Failed to select instance without critical clients")
	}
}

func TestDefaultComparatorCriticalResponseTime(t *testing.T) {
	c := stubClient()
	c.Config.Region = "A"
	c.Config.Host = "127.0.0.1"

	s1 := stubServiceInfo()
	s1.Config.Region = "A"
	s1.Config.CriticalAverageResponseTime = 2 * time.Second
	s1.Config.ServiceAddr.IPAddress = "192.168.1.1"
	s1.Stats.AverageResponseTime = 1 * time.Second

	s2 := stubServiceInfo()
	s2.Config.Region = "A"
	s2.Config.CriticalAverageResponseTime = 2 * time.Second
	s2.Config.ServiceAddr.IPAddress = "192.168.1.2"
	s2.Stats.AverageResponseTime = 5 * time.Second

	if !defaultComparator(c, s1, s2) || defaultComparator(c, s2, s1) {
		t.Error("Failed to select instance without critical response time")
	}

	s1.Stats.LastRequest = "2012-11-14T15:04:05Z-0700"
	s2.Stats.LastRequest = "2012-11-14T15:04:10Z-0700"

	// Tie breaker LastRequest
	if !defaultComparator(c, s1, s2) {
		t.Error("Failed to select instance with older LastRequest")
	}

	// Edge case, try both in different region
	s1.Config.Region = "B"
	s2.Config.Region = "B"

	if !defaultComparator(c, s1, s2) || defaultComparator(c, s2, s1) {
		t.Error("Failed to select instance without critical response time")
	}
}

func TestDefaultComparatorCriticalResponseTimeLeaveRegion(t *testing.T) {
	c := stubClient()
	c.Config.Region = "A"
	c.Config.Host = "127.0.0.1"

	s1 := stubServiceInfo()
	s1.Config.Region = "B"
	s1.Config.CriticalAverageResponseTime = 2 * time.Second
	s1.Config.ServiceAddr.IPAddress = "192.168.1.1"
	s1.Stats.AverageResponseTime = 1 * time.Second

	s2 := stubServiceInfo()
	s2.Config.Region = "A"
	s2.Config.CriticalAverageResponseTime = 2 * time.Second
	s2.Config.ServiceAddr.IPAddress = "192.168.1.2"
	s2.Stats.AverageResponseTime = 5 * time.Second

	if !defaultComparator(c, s1, s2) || defaultComparator(c, s2, s1) {
		t.Error("Failed to select instance without critical response time")
	}

	s1.Stats.LastRequest = "2012-11-14T15:04:05Z-0700"
	s2.Stats.LastRequest = "2012-11-14T15:04:10Z-0700"
}

func TestDefaultComparatorCriticalResponseTimeZeroValue(t *testing.T) {
	c := stubClient()
	c.Config.Region = "A"
	c.Config.Host = "127.0.0.1"

	s1 := stubServiceInfo()
	s1.Config.Region = "A"
	s1.Config.ServiceAddr.IPAddress = "192.168.1.1"
	s1.Stats.AverageResponseTime = 1 * time.Second

	s2 := stubServiceInfo()
	s2.Config.Region = "B"
	s2.Config.ServiceAddr.IPAddress = "192.168.1.2"
	s2.Stats.AverageResponseTime = 1 * time.Second

	if !defaultComparator(c, s1, s2) || defaultComparator(c, s2, s1) {
		t.Error("Failed to select instance without critical response time")
	}

	s1.Stats.LastRequest = "2012-11-14T15:04:05Z-0700"
	s2.Stats.LastRequest = "2012-11-14T15:04:10Z-0700"
}

func TestDefaultComparatorCriticalClientsBetterThanCriticalResponseTime(t *testing.T) {
	c := stubClient()
	c.Config.Region = "A"
	c.Config.Host = "127.0.0.1"

	s1 := stubServiceInfo()
	s1.Config.Region = "A"
	s1.Config.ServiceAddr.IPAddress = "192.168.1.1"
	s1.Config.CriticalAverageResponseTime = 2 * time.Second
	s1.Stats.AverageResponseTime = 1 * time.Second
	s1.Config.CriticalClientCount = 5
	s1.Stats.Clients = 6

	s2 := stubServiceInfo()
	s2.Config.Region = "A"
	s2.Config.ServiceAddr.IPAddress = "192.168.1.2"
	s2.Config.CriticalAverageResponseTime = 2 * time.Second
	s2.Stats.AverageResponseTime = 5 * time.Second
	s2.Config.CriticalClientCount = 5
	s2.Stats.Clients = 2

	if !defaultComparator(c, s1, s2) || defaultComparator(c, s2, s1) {
		t.Error("Failed to select instance without critical response time")
	}

	s1.Stats.LastRequest = "2012-11-14T15:04:05Z-0700"
	s2.Stats.LastRequest = "2012-11-14T15:04:10Z-0700"
}

func TestDefaultComparatorFullSort(t *testing.T) {
	c := stubClient()
	c.Config.Region = "A"
	c.Config.Host = "127.0.0.1"

	ic := NewInstanceChooser(c)
	instances := make([]*skynet.ServiceInfo, 16)

	// Same Host, no critical thresholds passed, last request
	instances[0] = stubServiceInfo()
	instances[0].Config.Region = "A"
	instances[0].Config.ServiceAddr.IPAddress = "127.0.0.1"
	instances[0].Config.ServiceAddr.Port = 1
	instances[0].Stats.LastRequest = "2012-11-14T15:04:05Z-0700"

	// Same Host, no critical thresholds passed
	instances[1] = stubServiceInfo()
	instances[1].Config.Region = "A"
	instances[1].Config.ServiceAddr.IPAddress = "127.0.0.1"
	instances[1].Config.ServiceAddr.Port = 2
	instances[1].Stats.LastRequest = "2012-11-14T15:04:09Z-0700"

	// Same Host, no critical thresholds passed
	instances[2] = stubServiceInfo()
	instances[2].Config.Region = "A"
	instances[2].Config.ServiceAddr.IPAddress = "127.0.0.1"
	instances[2].Config.ServiceAddr.Port = 3
	instances[2].Stats.LastRequest = "2012-11-14T15:04:12Z-0700"

	// Same Region, no critical thresholds passed, last request
	instances[3] = stubServiceInfo()
	instances[3].Config.Region = "A"
	instances[3].Config.ServiceAddr.IPAddress = "192.168.1.1"
	instances[3].Config.ServiceAddr.Port = 4
	instances[3].Stats.LastRequest = "2012-11-14T15:04:10Z-0700"

	// Same Region, no critical thresholds passed
	instances[4] = stubServiceInfo()
	instances[4].Config.Region = "A"
	instances[4].Config.ServiceAddr.IPAddress = "192.168.1.2"
	instances[4].Config.ServiceAddr.Port = 5
	instances[4].Stats.LastRequest = "2012-11-14T15:04:12Z-0700"

	// Different Region, no critical thresholds passed, last request
	instances[5] = stubServiceInfo()
	instances[5].Config.Region = "B"
	instances[5].Config.ServiceAddr.IPAddress = "192.168.1.1"
	instances[5].Config.ServiceAddr.Port = 6
	instances[5].Stats.LastRequest = "2012-11-14T15:04:10Z-0700"

	// Different Region, no critical thresholds passed
	instances[6] = stubServiceInfo()
	instances[6].Config.Region = "B"
	instances[6].Config.ServiceAddr.IPAddress = "192.168.1.2"
	instances[6].Config.ServiceAddr.Port = 7
	instances[6].Stats.LastRequest = "2012-11-14T15:04:12Z-0700"

	// Same Host, critical clients
	instances[7] = stubServiceInfo()
	instances[7].Config.Region = "A"
	instances[7].Config.ServiceAddr.IPAddress = "127.0.0.1"
	instances[7].Config.ServiceAddr.Port = 8
	instances[7].Config.CriticalClientCount = 5
	instances[7].Stats.Clients = 6

	// Same Host, critical response time
	instances[8] = stubServiceInfo()
	instances[8].Config.Region = "A"
	instances[8].Config.ServiceAddr.IPAddress = "127.0.0.1"
	instances[8].Config.ServiceAddr.Port = 9
	instances[8].Config.CriticalClientCount = 5
	instances[8].Stats.Clients = 2
	instances[8].Config.CriticalAverageResponseTime = 2 * time.Second
	instances[8].Stats.AverageResponseTime = 5 * time.Second

	// Same Region, critical clients
	instances[9] = stubServiceInfo()
	instances[9].Config.Region = "A"
	instances[9].Config.ServiceAddr.IPAddress = "192.168.1.3"
	instances[9].Config.ServiceAddr.Port = 10
	instances[9].Config.CriticalClientCount = 5
	instances[9].Stats.Clients = 6

	// Same Region, critical response time
	instances[10] = stubServiceInfo()
	instances[10].Config.Region = "A"
	instances[10].Config.ServiceAddr.IPAddress = "192.168.1.4"
	instances[10].Config.ServiceAddr.Port = 11
	instances[10].Config.CriticalClientCount = 5
	instances[10].Stats.Clients = 2
	instances[10].Config.CriticalAverageResponseTime = 2 * time.Second
	instances[10].Stats.AverageResponseTime = 5 * time.Second

	// Different Region, critical clients
	instances[11] = stubServiceInfo()
	instances[11].Config.Region = "B"
	instances[11].Config.ServiceAddr.IPAddress = "192.168.1.3"
	instances[11].Config.ServiceAddr.Port = 12
	instances[11].Config.CriticalClientCount = 5
	instances[11].Stats.Clients = 6

	// Different Region, critical response time
	instances[12] = stubServiceInfo()
	instances[12].Config.Region = "B"
	instances[12].Config.ServiceAddr.IPAddress = "192.168.1.4"
	instances[12].Config.ServiceAddr.Port = 13
	instances[12].Config.CriticalClientCount = 5
	instances[12].Stats.Clients = 2
	instances[12].Config.CriticalAverageResponseTime = 2 * time.Second
	instances[12].Stats.AverageResponseTime = 5 * time.Second

	// Same Host, critical response time, and critical clients
	instances[13] = stubServiceInfo()
	instances[13].Config.Region = "A"
	instances[13].Config.ServiceAddr.IPAddress = "127.0.0.1"
	instances[13].Config.ServiceAddr.Port = 14
	instances[13].Config.CriticalClientCount = 5
	instances[13].Stats.Clients = 7
	instances[13].Config.CriticalAverageResponseTime = 2 * time.Second
	instances[13].Stats.AverageResponseTime = 5 * time.Second

	// Same Region, critical response time, and critical clients
	instances[14] = stubServiceInfo()
	instances[14].Config.Region = "A"
	instances[14].Config.ServiceAddr.IPAddress = "192.168.1.5"
	instances[14].Config.ServiceAddr.Port = 15
	instances[14].Config.CriticalClientCount = 5
	instances[14].Stats.Clients = 7
	instances[14].Config.CriticalAverageResponseTime = 2 * time.Second
	instances[14].Stats.AverageResponseTime = 5 * time.Second

	// Different Region, critical response time, and critical clients
	instances[15] = stubServiceInfo()
	instances[15].Config.Region = "B"
	instances[15].Config.ServiceAddr.IPAddress = "192.168.1.5"
	instances[15].Config.ServiceAddr.Port = 16
	instances[15].Config.CriticalClientCount = 5
	instances[15].Stats.Clients = 7
	instances[15].Config.CriticalAverageResponseTime = 2 * time.Second
	instances[15].Stats.AverageResponseTime = 5 * time.Second

	// Add instances to InstanceChooser in random order to ensure sorting works
	ic.Add(instances[3])
	ic.Add(instances[6])
	ic.Add(instances[12])
	ic.Add(instances[9])
	ic.Add(instances[1])
	ic.Add(instances[15])
	ic.Add(instances[4])
	ic.Add(instances[13])
	ic.Add(instances[8])
	ic.Add(instances[11])
	ic.Add(instances[5])
	ic.Add(instances[0])
	ic.Add(instances[10])
	ic.Add(instances[7])
	ic.Add(instances[14])
	ic.Add(instances[2])

	// Don't need result, we need to look at the ordering of the array
	tc := make(chan bool, 1)

	/*
	 * Due to the way select{} works, it's possible that we get a race condition between Add and Choose
	 * in a production case we don't care because we will always get the highest priority instance that the
	 * instance chooser already knows about. For our test case we need to ensure that all Add's took place
	 * so that we can validate the order of the entire set.
	 */
	time.Sleep(250 * time.Millisecond)
	_, _ = ic.Choose(tc)

	for k, _ := range instances {
		if ic.instances[k] != instances[k] {
			t.Error("InstanceChooser did not sort properly, index: " + strconv.Itoa(k) + " is incorrect. Value: " + strconv.Itoa((ic.instances[k].Config.ServiceAddr.Port - 1)))
		}
	}
}
