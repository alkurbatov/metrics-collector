package monitoring

import (
	"runtime"

	"github.com/alkurbatov/metrics-collector/pkg/metrics"
)

// Metrics of Go runtime.
// Regarding puurpose of each value see:
// https://pkg.go.dev/runtime#MemStats
type RuntimeStats struct {
	Alloc         metrics.Gauge
	BuckHashSys   metrics.Gauge
	Frees         metrics.Gauge
	GCCPUFraction metrics.Gauge
	GCSys         metrics.Gauge
	HeapAlloc     metrics.Gauge
	HeapIdle      metrics.Gauge
	HeapInuse     metrics.Gauge
	HeapObjects   metrics.Gauge
	HeapReleased  metrics.Gauge
	HeapSys       metrics.Gauge
	LastGC        metrics.Gauge
	Lookups       metrics.Gauge
	MCacheInuse   metrics.Gauge
	MCacheSys     metrics.Gauge
	MSpanInuse    metrics.Gauge
	MSpanSys      metrics.Gauge
	Mallocs       metrics.Gauge
	NextGC        metrics.Gauge
	NumForcedGC   metrics.Gauge
	NumGC         metrics.Gauge
	OtherSys      metrics.Gauge
	PauseTotalNs  metrics.Gauge
	StackInuse    metrics.Gauge
	StackSys      metrics.Gauge
	Sys           metrics.Gauge
	TotalAlloc    metrics.Gauge
}

// Poll refreshes values of runtime metrics.
func (m *RuntimeStats) Poll() {
	stats := runtime.MemStats{}
	runtime.ReadMemStats(&stats)

	m.Alloc = metrics.Gauge(stats.Alloc)
	m.BuckHashSys = metrics.Gauge(stats.BuckHashSys)
	m.Frees = metrics.Gauge(stats.Frees)
	m.GCCPUFraction = metrics.Gauge(stats.GCCPUFraction)
	m.GCSys = metrics.Gauge(stats.GCSys)
	m.HeapAlloc = metrics.Gauge(stats.HeapAlloc)
	m.HeapIdle = metrics.Gauge(stats.HeapIdle)
	m.HeapInuse = metrics.Gauge(stats.HeapInuse)
	m.HeapObjects = metrics.Gauge(stats.HeapObjects)
	m.HeapReleased = metrics.Gauge(stats.HeapReleased)
	m.HeapSys = metrics.Gauge(stats.HeapSys)
	m.LastGC = metrics.Gauge(stats.LastGC)
	m.Lookups = metrics.Gauge(stats.Lookups)
	m.MCacheInuse = metrics.Gauge(stats.MCacheInuse)
	m.MCacheSys = metrics.Gauge(stats.MCacheSys)
	m.MSpanInuse = metrics.Gauge(stats.MSpanInuse)
	m.MSpanSys = metrics.Gauge(stats.MSpanSys)
	m.Mallocs = metrics.Gauge(stats.Mallocs)
	m.NextGC = metrics.Gauge(stats.NextGC)
	m.NumForcedGC = metrics.Gauge(stats.NumForcedGC)
	m.NumGC = metrics.Gauge(stats.NumGC)
	m.OtherSys = metrics.Gauge(stats.OtherSys)
	m.PauseTotalNs = metrics.Gauge(stats.PauseTotalNs)
	m.StackInuse = metrics.Gauge(stats.StackInuse)
	m.StackSys = metrics.Gauge(stats.StackSys)
	m.Sys = metrics.Gauge(stats.Sys)
	m.TotalAlloc = metrics.Gauge(stats.TotalAlloc)
}
