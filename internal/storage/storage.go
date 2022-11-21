package storage

import "github.com/alkurbatov/metrics-collector/internal/metrics"

type Storage interface {
	PushCounter(name string, value metrics.Counter)
	PushGauge(name string, value metrics.Gauge)
}
