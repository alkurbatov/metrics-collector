package services

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/alkurbatov/metrics-collector/internal/storage"
)

type MetricsRecorder struct {
	storage storage.Storage
}

func NewMetricsRecorder(dataStore storage.Storage) MetricsRecorder {
	return MetricsRecorder{storage: dataStore}
}

func (r MetricsRecorder) PushCounter(ctx context.Context, name string, value metrics.Counter) (metrics.Counter, error) {
	id := name + "_counter"

	prevValue, err := r.storage.Get(ctx, id)
	if err != nil && !errors.Is(err, entity.ErrMetricNotFound) {
		return 0, err
	}

	if err == nil {
		value += prevValue.Value.(metrics.Counter)
	}

	err = r.storage.Push(ctx, id, storage.Record{Name: name, Value: value})
	if err != nil {
		return 0, err
	}

	return value, nil
}

func (r MetricsRecorder) PushGauge(ctx context.Context, name string, value metrics.Gauge) (metrics.Gauge, error) {
	id := name + "_gauge"

	err := r.storage.Push(ctx, id, storage.Record{Name: name, Value: value})
	if err != nil {
		return 0, err
	}

	return value, nil
}

func (r MetricsRecorder) GetRecord(ctx context.Context, kind, name string) (*storage.Record, error) {
	id := fmt.Sprintf("%s_%s", name, kind)

	return r.storage.Get(ctx, id)
}

func (r MetricsRecorder) ListRecords(ctx context.Context) ([]storage.Record, error) {
	values, err := r.storage.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	rv := append([]storage.Record(nil), values...)

	sort.Slice(rv, func(i, j int) bool {
		return rv[i].Name < rv[j].Name
	})

	return rv, nil
}
