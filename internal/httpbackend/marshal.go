package httpbackend

import (
	"context"
	"fmt"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/alkurbatov/metrics-collector/internal/validators"
	"github.com/alkurbatov/metrics-collector/pkg/metrics"
	"github.com/rs/zerolog/log"
)

func toRecord(ctx context.Context, req *metrics.MetricReq, signer *security.Signer) (storage.Record, error) {
	var record storage.Record

	if err := validators.ValidateMetricName(req.ID, req.MType); err != nil {
		return record, err
	}

	switch req.MType {
	case metrics.KindCounter:
		if req.Delta == nil {
			return record, entity.ErrIncompleteRequest
		}

		record = storage.Record{Name: req.ID, Value: *req.Delta}

	case metrics.KindGauge:
		if req.Value == nil {
			return record, entity.ErrIncompleteRequest
		}

		record = storage.Record{Name: req.ID, Value: *req.Value}

	default:
		return record, entity.MetricNotImplementedError(req.MType)
	}

	if signer != nil {
		valid, err := signer.VerifyRecordSignature(record, req.Hash)
		if err != nil {
			// NB (alkurbatov): We don't want to give any hints to potential attacker,
			// but still want to debug implementation errors. Thus, the error is only logged.
			log.Ctx(ctx).Error().Err(err).Msg("")
		}

		if err != nil || !valid {
			return record, entity.ErrInvalidSignature
		}
	}

	return record, nil
}

func toMetricReq(record storage.Record, signer *security.Signer) (*metrics.MetricReq, error) {
	req := &metrics.MetricReq{ID: record.Name, MType: record.Value.Kind()}

	if signer != nil {
		hash, err := signer.CalculateRecordSignature(record)
		if err != nil {
			return nil, fmt.Errorf("httpbackend - toMetricReq - signer.CalculateRecordSignature: %w", err)
		}

		req.Hash = hash
	}

	switch record.Value.Kind() {
	case metrics.KindCounter:
		delta, _ := record.Value.(metrics.Counter)
		req.Delta = &delta

	case metrics.KindGauge:
		value, _ := record.Value.(metrics.Gauge)
		req.Value = &value
	}

	return req, nil
}

func toMetricReqList(records []storage.Record, signer *security.Signer) ([]*metrics.MetricReq, error) {
	rv := make([]*metrics.MetricReq, len(records))

	for i, record := range records {
		req, err := toMetricReq(record, signer)
		if err != nil {
			return nil, fmt.Errorf("httpbackend - toMetricReqList - toMetricReq: %w", err)
		}

		rv[i] = req
	}

	return rv, nil
}
