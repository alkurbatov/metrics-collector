package services

import (
	"context"
	"errors"
	"sort"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/alkurbatov/metrics-collector/internal/storage"
)

func calculateID(name, kind string) string {
	return name + "_" + kind
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
		return nil, err
	}

	return storedRecord.Value.(metrics.Counter) + newRecord.Value.(metrics.Counter), nil
}

func (r MetricsRecorder) Push(ctx context.Context, record storage.Record) (storage.Record, error) {
	id := calculateID(record.Name, record.Value.Kind())

	value, err := r.calculateNewValue(ctx, id, nil, record)
	if err != nil {
		return storage.Record{}, err
	}

	record.Value = value
	if err := r.storage.Push(ctx, id, record); err != nil {
		return storage.Record{}, err
	}

	return record, nil
}

func (r MetricsRecorder) PushList(ctx context.Context, records []storage.Record) error {
	keys := make([]string, 0)
	data := make([]storage.Record, 0)
	seen := make(map[string]int)

	for _, record := range records {
		id := calculateID(record.Name, record.Value.Kind())

		// NB (alkurbatov): Compress requests to metrics with same names.
		if pos, ok := seen[id]; ok {
			value, err := r.calculateNewValue(ctx, id, &data[pos], record)
			if err != nil {
				return err
			}

			data[pos].Value = value

			continue
		}

		value, err := r.calculateNewValue(ctx, id, nil, record)
		if err != nil {
			return err
		}

		record.Value = value
		seen[id] = len(data)

		keys = append(keys, id)
		data = append(data, record)
	}

	return r.storage.PushList(ctx, keys, data)
}

func (r MetricsRecorder) Get(ctx context.Context, kind, name string) (storage.Record, error) {
	id := calculateID(name, kind)

	return r.storage.Get(ctx, id)
}

func (r MetricsRecorder) List(ctx context.Context) ([]storage.Record, error) {
	rv, err := r.storage.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	sort.Slice(rv, func(i, j int) bool {
		return rv[i].Name < rv[j].Name
	})

	return rv, nil
}
