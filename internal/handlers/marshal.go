package handlers

import (
	"context"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/logging"
	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/alkurbatov/metrics-collector/internal/schema"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/alkurbatov/metrics-collector/internal/storage"
)

func toRecord(ctx context.Context, req *schema.MetricReq, signer *security.Signer) (storage.Record, error) {
	if err := schema.ValidateMetricName(req.ID, req.MType); err != nil {
		return storage.Record{}, err
	}

	if signer != nil {
		valid, err := signer.VerifySignature(req)
		if err != nil {
			// NB (alkurbatov): We don't want to give any hints to potential attacker,
			// but still want to debug implementation errors. Thus, the error is only logged.
			logging.GetLogger(ctx).Error().Err(err).Msg("")
		}

		if err != nil || !valid {
			return storage.Record{}, entity.ErrInvalidSignature
		}
	}

	switch req.MType {
	case "counter":
		if req.Delta == nil {
			return storage.Record{}, entity.ErrIncompleteRequest
		}

		return storage.Record{Name: req.ID, Value: *req.Delta}, nil

	case "gauge":
		if req.Value == nil {
			return storage.Record{}, entity.ErrIncompleteRequest
		}

		return storage.Record{Name: req.ID, Value: *req.Value}, nil

	default:
		return storage.Record{}, entity.MetricNotImplementedError(req.MType)
	}
}

func toMetricReq(record storage.Record) schema.MetricReq {
	req := schema.MetricReq{ID: record.Name, MType: record.Value.Kind()}

	switch record.Value.Kind() {
	case entity.Counter:
		delta, _ := record.Value.(metrics.Counter)
		req.Delta = &delta

	case entity.Gauge:
		value, _ := record.Value.(metrics.Gauge)
		req.Value = &value
	}

	return req
}
