package metrics

import "runtime"

// Regarding puurpose of each value see:
// https://pkg.go.dev/runtime#MemStats
type RuntimeStats struct {
	Alloc         Gauge
	BuckHashSys   Gauge
	Frees         Gauge
	GCCPUFraction Gauge
	GCSys         Gauge
	HeapAlloc     Gauge
	HeapIdle      Gauge
	HeapInuse     Gauge
	HeapObjects   Gauge
	HeapReleased  Gauge
	HeapSys       Gauge
	LastGC        Gauge
	Lookups       Gauge
	MCacheInuse   Gauge
	MCacheSys     Gauge
	MSpanInuse    Gauge
	MSpanSys      Gauge
	Mallocs       Gauge
	NextGC        Gauge
	NumForcedGC   Gauge
	NumGC         Gauge
	OtherSys      Gauge
	PauseTotalNs  Gauge
	StackInuse    Gauge
	StackSys      Gauge
	Sys           Gauge
	TotalAlloc    Gauge
}

func (m *RuntimeStats) Poll() {
	stats := runtime.MemStats{}
	runtime.ReadMemStats(&stats)

	m.Alloc = Gauge(stats.Alloc)
	m.BuckHashSys = Gauge(stats.BuckHashSys)
	m.Frees = Gauge(stats.Frees)
	m.GCCPUFraction = Gauge(stats.GCCPUFraction)
	m.GCSys = Gauge(stats.GCSys)
	m.HeapAlloc = Gauge(stats.HeapAlloc)
	m.HeapIdle = Gauge(stats.HeapIdle)
	m.HeapInuse = Gauge(stats.HeapInuse)
	m.HeapObjects = Gauge(stats.HeapObjects)
	m.HeapReleased = Gauge(stats.HeapReleased)
	m.HeapSys = Gauge(stats.HeapSys)
	m.LastGC = Gauge(stats.LastGC)
	m.Lookups = Gauge(stats.Lookups)
	m.MCacheInuse = Gauge(stats.MCacheInuse)
	m.MCacheSys = Gauge(stats.MCacheSys)
	m.MSpanInuse = Gauge(stats.MSpanInuse)
	m.MSpanSys = Gauge(stats.MSpanSys)
	m.Mallocs = Gauge(stats.Mallocs)
	m.NextGC = Gauge(stats.NextGC)
	m.NumForcedGC = Gauge(stats.NumForcedGC)
	m.NumGC = Gauge(stats.NumGC)
	m.OtherSys = Gauge(stats.OtherSys)
	m.PauseTotalNs = Gauge(stats.PauseTotalNs)
	m.StackInuse = Gauge(stats.StackInuse)
	m.StackSys = Gauge(stats.StackSys)
	m.Sys = Gauge(stats.Sys)
	m.TotalAlloc = Gauge(stats.TotalAlloc)
}
