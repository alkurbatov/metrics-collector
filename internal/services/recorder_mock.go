package services

import (
	"errors"

	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/alkurbatov/metrics-collector/internal/storage"
)

type RecorderMock struct {
}

func (m RecorderMock) PushCounter(name string, value metrics.Counter) (metrics.Counter, error) {
	if name == "fail" {
		return 0, errors.New("failure")
	}

	return value, nil
}

func (m RecorderMock) PushGauge(name string, value metrics.Gauge) (metrics.Gauge, error) {
	if name == "fail" {
		return 0, errors.New("failure")
	}

	return value, nil
}

func (m RecorderMock) GetRecord(kind, name string) (storage.Record, bool) {
	if name == "unknown" {
		return storage.Record{}, false
	}

	switch kind {
	case "counter":
		return storage.Record{Name: name, Value: metrics.Counter(10)}, true
	case "gauge":
		return storage.Record{Name: name, Value: metrics.Gauge(11.345)}, true
	default:
		return storage.Record{}, false
	}
}

func (m RecorderMock) ListRecords() []storage.Record {
	rv := []storage.Record{
		{Name: "A", Value: metrics.Counter(10)},
		{Name: "B", Value: metrics.Gauge(11.345)},
	}

	return rv
}
