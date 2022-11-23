package storage

import (
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/stretchr/testify/assert"
)

func TestPushCounter(t *testing.T) {
	m := NewMemStorage()
	name := "PollCount"
	assert := assert.New(t)

	m.PushCounter(name, 10)
	assert.Equal(metrics.Counter(10), m.counters[name])

	m.PushCounter(name, 23)
	assert.Equal(metrics.Counter(33), m.counters[name])
}

func TestPushGauge(t *testing.T) {
	m := NewMemStorage()
	name := "Alloc"
	assert := assert.New(t)

	m.PushGauge(name, 10.1234)
	assert.Equal(metrics.Gauge(10.1234), m.gauges[name])

	m.PushGauge(name, 0.567)
	assert.Equal(metrics.Gauge(0.567), m.gauges[name])
}
