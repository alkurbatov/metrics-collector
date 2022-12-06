package services

import (
	"fmt"
	"sort"

	"github.com/alkurbatov/metrics-collector/internal/app"
	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/alkurbatov/metrics-collector/internal/storage"
)

type MetricsRecorder struct {
	storage storage.Storage
}

func NewMetricsRecorder(app *app.Server) MetricsRecorder {
	return MetricsRecorder{storage: app.Storage}
}

func (r MetricsRecorder) PushCounter(name, rawValue string) error {
	value, err := metrics.ToCounter(rawValue)
	if err != nil {
		return err
	}

	id := name + "_counter"

	prevValue, ok := r.storage.Get(id)
	if ok {
		value += prevValue.Value.(metrics.Counter)
	}

	r.storage.Push(id, storage.Record{Name: name, Value: value})
	return nil
}

func (r MetricsRecorder) PushGauge(name, rawValue string) error {
	value, err := metrics.ToGauge(rawValue)
	if err != nil {
		return err
	}

	id := name + "_gauge"

	r.storage.Push(id, storage.Record{Name: name, Value: value})
	return nil
}

func (r MetricsRecorder) GetRecord(kind, name string) (storage.Record, bool) {
	id := fmt.Sprintf("%s_%s", name, kind)

	return r.storage.Get(id)
}

func (r MetricsRecorder) ListRecords() []storage.Record {
	rv := append([]storage.Record(nil), r.storage.GetAll()...)

	sort.Slice(rv, func(i, j int) bool {
		return rv[i].Name < rv[j].Name
	})

	return rv
}
