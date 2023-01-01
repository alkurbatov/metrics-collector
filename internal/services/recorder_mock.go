package services

import (
	"context"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/alkurbatov/metrics-collector/internal/storage"
)

type RecorderMock struct {
}

func (m RecorderMock) PushCounter(ctx context.Context, name string, value metrics.Counter) (metrics.Counter, error) {
	if name == "fail" {
		return 0, entity.ErrUnexpected
	}

	return value, nil
}

func (m RecorderMock) PushGauge(ctx context.Context, name string, value metrics.Gauge) (metrics.Gauge, error) {
	if name == "fail" {
		return 0, entity.ErrUnexpected
	}

	return value, nil
}

func (m RecorderMock) GetRecord(ctx context.Context, kind, name string) (*storage.Record, error) {
	if name == "unknown" {
		return nil, entity.ErrMetricNotFound
	}

	switch kind {
	case entity.Counter:
		return &storage.Record{Name: name, Value: metrics.Counter(10)}, nil
	case entity.Gauge:
		return &storage.Record{Name: name, Value: metrics.Gauge(11.345)}, nil
	default:
		return nil, entity.ErrMetricNotImplemented
	}
}

func (m RecorderMock) ListRecords(ctx context.Context) ([]storage.Record, error) {
	rv := []storage.Record{
		{Name: "A", Value: metrics.Counter(10)},
		{Name: "B", Value: metrics.Gauge(11.345)},
	}

	return rv, nil
}
