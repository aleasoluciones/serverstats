package serverstats

import (
	"fmt"
	"os"
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
	Timestamp int64  `json:"timestamp"`
	Name      string `json:"name"`
	Value     string `json:"value"`
	Unit      string `json:"unit,omitempty"`
}

func loadavg(cs chan Metric, hostname string) {
	now := time.Now().Unix()
	concreteSigar := sigar.ConcreteSigar{}

	avg, err := concreteSigar.GetLoadAverage()
	if err != nil {
		fmt.Printf("Failed to get load average")
		return
	}

	cs <- createMetric(now, hostname, "loadavg.one", fmt.Sprintf("%.2f", avg.One), "")
	cs <- createMetric(now, hostname, "loadavg.five", fmt.Sprintf("%.2f", avg.Five), "")
	cs <- createMetric(now, hostname, "loadavg.Fifteen", fmt.Sprintf("%.2f", avg.Five), "")
}

func memStats(cs chan Metric, hostname string) {
	now := time.Now().Unix()
	mem := sigar.Mem{}
	swap := sigar.Swap{}

	mem.Get()
	swap.Get()

	cs <- createMetric(now, hostname, "mem.total", toMegabytesString(mem.Total), "M")
	cs <- createMetric(now, hostname, "mem.used", toMegabytesString(mem.Used), "M")
	cs <- createMetric(now, hostname, "mem.free", toMegabytesString(mem.Free), "M")
	cs <- createMetric(now, hostname, "mem.actualused", toMegabytesString(mem.ActualUsed), "M")
	cs <- createMetric(now, hostname, "mem.actualfree", toMegabytesString(mem.ActualFree), "M")
	cs <- createMetric(now, hostname, "swap.total", toMegabytesString(swap.Total), "M")
	cs <- createMetric(now, hostname, "swap.used", toMegabytesString(swap.Used), "M")
	cs <- createMetric(now, hostname, "swap.free", toMegabytesString(swap.Free), "M")

	memActualUsedPercent := (float64(mem.ActualUsed) / float64(mem.Total)) * 100
	swapUsedPercent := (float64(swap.Used) / float64(swap.Total)) * 100
	cs <- createMetric(now, hostname, "mem.actualusedpercent", fmt.Sprintf("%.2f", memActualUsedPercent), "%")
	cs <- createMetric(now, hostname, "swap.usedpercent", fmt.Sprintf("%.2f", swapUsedPercent), "%")
}

func toMegabytesString(bytes uint64) string {
	return fmt.Sprintf("%.2f", (float64(bytes) / 1024 / 1024))
}

func cpuStatsLoop(ch chan Metric, hostname string, periode time.Duration) {
	concreteSigar := sigar.ConcreteSigar{}
	cpuCh, _ := concreteSigar.CollectCpuStats(periode)
	for cpuStat := range cpuCh {
		now := time.Now().Unix()
		total := float64(cpuStat.Total())

		user := (float64(cpuStat.User) / total) * 100
		sys := (float64(cpuStat.Sys) / total) * 100
		idle := (float64(cpuStat.Idle) / total) * 100
		wait := (float64(cpuStat.Wait) / total) * 100
		stolen := (float64(cpuStat.Stolen) / total) * 100

		ch <- createMetric(now, hostname, "cpu.user", fmt.Sprintf("%.2f", user), "%")
		ch <- createMetric(now, hostname, "cpu.sys", fmt.Sprintf("%.2f", sys), "%")
		ch <- createMetric(now, hostname, "cpu.idle", fmt.Sprintf("%.2f", idle), "%")
		ch <- createMetric(now, hostname, "cpu.wait", fmt.Sprintf("%.2f", wait), "%")
		ch <- createMetric(now, hostname, "cpu.stolen", fmt.Sprintf("%.2f", stolen), "%")
	}
}

func createMetric(timestamp int64, hostname, name, value string, unit string) Metric {
	return Metric{
		Timestamp: timestamp,
		Name:      hostname + "." + name,
		Value:     value,
		Unit:      unit,
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
	hostname, _ := os.Hostname()
	scheduledtask.NewScheduledTask(func() { memStats(serverStats.Metrics, hostname) }, periodes.Mem, 0)
	scheduledtask.NewScheduledTask(func() { loadavg(serverStats.Metrics, hostname) }, periodes.LoadAvg, 0)
	go cpuStatsLoop(serverStats.Metrics, hostname, periodes.Cpu)
	return &serverStats
}
