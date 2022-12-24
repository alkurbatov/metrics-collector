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
	m.PollCount += 1
	m.RandomValue = Gauge(rand.Float64())
	m.Memory.Poll()
}
