package grpcbackend

import (
	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/alkurbatov/metrics-collector/internal/validators"
	"github.com/alkurbatov/metrics-collector/pkg/grpcapi"
	"github.com/alkurbatov/metrics-collector/pkg/metrics"
)

func toRecord(req *grpcapi.MetricReq) (storage.Record, error) {
	if err := validators.ValidateMetricName(req.Id, req.Mtype); err != nil {
		return storage.Record{}, err //nolint: wrapcheck
	}

	switch req.Mtype {
	case metrics.KindCounter:
		return storage.Record{Name: req.Id, Value: metrics.Counter(req.Delta)}, nil

	case metrics.KindGauge:
		return storage.Record{Name: req.Id, Value: metrics.Gauge(req.Value)}, nil

	default:
		return storage.Record{}, entity.MetricNotImplementedError(req.Mtype) //nolint: wrapcheck
	}
}

func toMetricReq(record storage.Record) *grpcapi.MetricReq {
	req := &grpcapi.MetricReq{Id: record.Name, Mtype: record.Value.Kind()}

	switch record.Value.Kind() {
	case metrics.KindCounter:
		delta, _ := record.Value.(metrics.Counter)
		req.Delta = int64(delta)

	case metrics.KindGauge:
		value, _ := record.Value.(metrics.Gauge)
		req.Value = float64(value)
	}

	return req
}
