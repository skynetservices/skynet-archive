package metrics

import (
	"github.com/jondot/gosigar"
)

type HostStats struct {
	Uptime      sigar.Uptime
	LoadAverage sigar.LoadAverage
	Mem         sigar.Mem
	Swap        sigar.Swap
	Cpu         sigar.Cpu
}

func (hs *HostStats) Update() {
	hs.Uptime.Get()
	hs.LoadAverage.Get()
	hs.Mem.Get()
	hs.Swap.Get()
	hs.Cpu.Get()
}
