package metrics

import (
	"math/rand"
)

type Metrics struct {
	Memory MemoryStats

	RandomValue Gauge

	PollCount Counter
}

func (m *Metrics) Poll() {
	m.PollCount++
	m.RandomValue = Gauge(rand.Float64()) //nolint: gosec
	m.Memory.Poll()
}
