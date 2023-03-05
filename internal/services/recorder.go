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

func CalculateID(name, kind string) string {
	return name + "_" + kind
}

func pushError(reason error) error {
	return fmt.Errorf("failed to push record: %w", reason)
}

func pushListError(reason error) error {
	return fmt.Errorf("failed to push records list: %w", reason)
}

type MetricsRecorder struct {
	storage storage.Storage
}

func NewMetricsRecorder(dataStore storage.Storage) MetricsRecorder {
	return MetricsRecorder{storage: dataStore}
}

func (r MetricsRecorder) calculateNewValue(
	ctx context.Context,
	key string,
	prevRecord *storage.Record,
	newRecord storage.Record,
) (metrics.Metric, error) {
	if newRecord.Value.Kind() != entity.Counter {
		return newRecord.Value, nil
	}

	if prevRecord != nil {
		return prevRecord.Value.(metrics.Counter) + newRecord.Value.(metrics.Counter), nil
	}

	storedRecord, err := r.storage.Get(ctx, key)
	if errors.Is(err, entity.ErrMetricNotFound) {
		return newRecord.Value, nil
	}

	if err != nil {
		return nil, err //nolint: wrapcheck
	}

	return storedRecord.Value.(metrics.Counter) + newRecord.Value.(metrics.Counter), nil
}

func (r MetricsRecorder) Push(ctx context.Context, record storage.Record) (storage.Record, error) {
	id := CalculateID(record.Name, record.Value.Kind())

	value, err := r.calculateNewValue(ctx, id, nil, record)
	if err != nil {
		return storage.Record{}, pushError(err)
	}

	record.Value = value
	if err := r.storage.Push(ctx, id, record); err != nil {
		return storage.Record{}, pushError(err)
	}

	return record, nil
}

func (r MetricsRecorder) PushList(ctx context.Context, records []storage.Record) error {
	data := make(map[string]storage.Record)

	for _, record := range records {
		id := CalculateID(record.Name, record.Value.Kind())

		if prev, ok := data[id]; ok {
			// NB (alkurbatov): Compress metrics with same names.
			value, err := r.calculateNewValue(ctx, id, &prev, record)
			if err != nil {
				return pushListError(err)
			}

			prev.Value = value
			data[id] = prev

			continue
		}

		value, err := r.calculateNewValue(ctx, id, nil, record)
		if err != nil {
			return pushListError(err)
		}

		record.Value = value
		data[id] = record
	}

	if err := r.storage.PushBatch(ctx, data); err != nil {
		return pushListError(err)
	}

	return nil
}

func (r MetricsRecorder) Get(ctx context.Context, kind, name string) (storage.Record, error) {
	id := CalculateID(name, kind)

	record, err := r.storage.Get(ctx, id)
	if err != nil {
		return storage.Record{}, fmt.Errorf("failed to get record: %w", err)
	}

	return record, nil
}

func (r MetricsRecorder) List(ctx context.Context) ([]storage.Record, error) {
	rv, err := r.storage.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list records: %w", err)
	}

	sort.Slice(rv, func(i, j int) bool {
		return rv[i].Name < rv[j].Name
	})

	return rv, nil
}
