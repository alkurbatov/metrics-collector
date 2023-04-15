package httpbackend

import (
	"context"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/alkurbatov/metrics-collector/pkg/metrics"
	"github.com/rs/zerolog/log"
)

func toRecord(ctx context.Context, req *metrics.MetricReq, signer *security.Signer) (storage.Record, error) {
	if err := ValidateMetricName(req.ID, req.MType); err != nil {
		return storage.Record{}, err
	}

	if signer != nil {
		valid, err := signer.VerifySignature(req)
		if err != nil {
			// NB (alkurbatov): We don't want to give any hints to potential attacker,
			// but still want to debug implementation errors. Thus, the error is only logged.
			log.Ctx(ctx).Error().Err(err).Msg("")
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
		return storage.Record{}, entity.MetricNotImplementedError(req.MType) //nolint: wrapcheck
	}
}

func toMetricReq(record storage.Record) metrics.MetricReq {
	req := metrics.MetricReq{ID: record.Name, MType: record.Value.Kind()}

	switch record.Value.Kind() {
	case metrics.KindCounter:
		delta, _ := record.Value.(metrics.Counter)
		req.Delta = &delta

	case metrics.KindGauge:
		value, _ := record.Value.(metrics.Gauge)
		req.Value = &value
	}

	return req
}
