package metrics

import (
	"math/rand"
)

type Metrics struct {
	Memory MemoryStats

	RandomValue Gauge

	PollCount Counter
}

func NewMetrics() *Metrics {
	return &Metrics{RandomValue: Gauge(rand.Float64())}
}

func (m *Metrics) Poll() {
	m.PollCount += 1
	m.Memory.Poll()
}
