package serverstats

import (
	"fmt"
	"time"

	"github.com/aleasoluciones/goaleasoluciones/scheduledtask"
	"github.com/cloudfoundry/gosigar"
)

const (
	MEMSTATS_PERIODE           = 10 * time.Second
	LOADAVG_AND_UPTIME_PERIODE = 10 * time.Second
	CPU_PERIODE                = 5 * time.Second
)

type Metric struct {
	Time  time.Time
	Name  string
	Value string
	Unit  string
}

func loadavg(cs chan Metric) {
	now := time.Now()
	concreteSigar := sigar.ConcreteSigar{}

	avg, err := concreteSigar.GetLoadAverage()
	if err != nil {
		fmt.Printf("Failed to get load average")
		return
	}

	cs <- Metric{Time: now, Name: "loadavg.one", Value: fmt.Sprintf("%.2f", avg.One)}
	cs <- Metric{Time: now, Name: "loadavg.five", Value: fmt.Sprintf("%.2f", avg.Five)}
	cs <- Metric{Time: now, Name: "loadavg.Fifteen", Value: fmt.Sprintf("%.2f", avg.Five)}
}

func toMegabytesString(bytes uint64) string {
	return fmt.Sprintf("%.2f", (float64(bytes) / 1024 / 1024))
}

func memStats(cs chan Metric) {
	now := time.Now()
	mem := sigar.Mem{}
	swap := sigar.Swap{}

	mem.Get()
	swap.Get()

	cs <- Metric{Time: now, Name: "mem.total", Value: toMegabytesString(mem.Total), Unit: "M"}
	cs <- Metric{Time: now, Name: "mem.used", Value: toMegabytesString(mem.Used), Unit: "M"}
	cs <- Metric{Time: now, Name: "mem.free", Value: toMegabytesString(mem.Free), Unit: "M"}
	cs <- Metric{Time: now, Name: "mem.actualused", Value: toMegabytesString(mem.ActualUsed), Unit: "M"}
	cs <- Metric{Time: now, Name: "mem.actualfree", Value: toMegabytesString(mem.ActualFree), Unit: "M"}
	cs <- Metric{Time: now, Name: "swap.total", Value: toMegabytesString(swap.Total), Unit: "M"}
	cs <- Metric{Time: now, Name: "swap.used", Value: toMegabytesString(swap.Used), Unit: "M"}
	cs <- Metric{Time: now, Name: "swap.free", Value: toMegabytesString(swap.Free), Unit: "M"}
}

func cpuStatsLoop(ch chan Metric, periode time.Duration) {

	concreteSigar := sigar.ConcreteSigar{}
	cpuCh, _ := concreteSigar.CollectCpuStats(periode)
	for cpuStat := range cpuCh {
		now := time.Now()
		total := float64(cpuStat.Total())

		user := (float64(cpuStat.User) / total) * 100
		sys := (float64(cpuStat.Sys) / total) * 100
		idle := (float64(cpuStat.Idle) / total) * 100
		wait := (float64(cpuStat.Wait) / total) * 100
		stolen := (float64(cpuStat.Stolen) / total) * 100

		ch <- Metric{Time: now, Name: "cpu.user", Value: fmt.Sprintf("%.2f", user), Unit: "%"}
		ch <- Metric{Time: now, Name: "cpu.sys", Value: fmt.Sprintf("%.2f", sys), Unit: "%"}
		ch <- Metric{Time: now, Name: "cpu.idle", Value: fmt.Sprintf("%.2f", idle), Unit: "%"}
		ch <- Metric{Time: now, Name: "cpu.wait", Value: fmt.Sprintf("%.2f", wait), Unit: "%"}
		ch <- Metric{Time: now, Name: "cpu.stolen", Value: fmt.Sprintf("%.2f", stolen), Unit: "%"}
	}
}

type ServerStatsPeriodes struct {
	Mem     time.Duration
	LoadAvg time.Duration
	Cpu     time.Duration
}

var DefaultPeriodes = &ServerStatsPeriodes{
	Mem:     1 * time.Second,
	LoadAvg: 1 * time.Second,
	Cpu:     1 * time.Second,
}

type ServerStats struct {
	Metrics chan Metric
}

func NewServerStats(periodes *ServerStatsPeriodes) *ServerStats {
	serverStats := ServerStats{
		Metrics: make(chan Metric),
	}
	scheduledtask.NewScheduledTask(func() { memStats(serverStats.Metrics) }, periodes.Mem, 0)
	scheduledtask.NewScheduledTask(func() { loadavg(serverStats.Metrics) }, periodes.LoadAvg, 0)
	go cpuStatsLoop(serverStats.Metrics, periodes.Cpu)
	return &serverStats
}
