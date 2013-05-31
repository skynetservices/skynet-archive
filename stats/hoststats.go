package stats

import (
	"github.com/jondot/gosigar"
)

type Host struct {
	Uptime      sigar.Uptime
	LoadAverage sigar.LoadAverage
	Mem         sigar.Mem
	Swap        sigar.Swap
	Cpu         sigar.Cpu
}

func (h *Host) Update() {
	h.Uptime.Get()
	h.LoadAverage.Get()
	h.Mem.Get()
	h.Swap.Get()
	h.Cpu.Get()
}
