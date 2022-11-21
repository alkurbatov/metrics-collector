package storage

import "github.com/alkurbatov/metrics-collector/internal/metrics"

type MemStorage struct {
	counters map[string]metrics.Counter
	gauges   map[string]metrics.Gauge
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		counters: make(map[string]metrics.Counter),
		gauges:   make(map[string]metrics.Gauge),
	}
}

func (m *MemStorage) PushCounter(name string, value metrics.Counter) {
	m.counters[name] += value
}

func (m *MemStorage) PushGauge(name string, value metrics.Gauge) {
	m.gauges[name] = value
}
