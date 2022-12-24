package services

import (
	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/alkurbatov/metrics-collector/internal/storage"
)

type Recorder interface {
	PushCounter(name string, value metrics.Counter) (metrics.Counter, error)
	PushGauge(name string, value metrics.Gauge) (metrics.Gauge, error)
	GetRecord(kind, name string) (storage.Record, bool)
	ListRecords() []storage.Record
}
