package services

import (
	"context"

	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/alkurbatov/metrics-collector/internal/storage"
)

type Recorder interface {
	PushCounter(ctx context.Context, name string, value metrics.Counter) (metrics.Counter, error)
	PushGauge(ctx context.Context, name string, value metrics.Gauge) (metrics.Gauge, error)
	GetRecord(ctx context.Context, kind, name string) (*storage.Record, error)
	ListRecords(ctx context.Context) ([]storage.Record, error)
}

type HealthCheck interface {
	CheckStorage(ctx context.Context) error
}
