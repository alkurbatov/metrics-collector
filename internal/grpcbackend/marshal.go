package grpcbackend

import (
	"context"
	"fmt"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/alkurbatov/metrics-collector/internal/validators"
	"github.com/alkurbatov/metrics-collector/pkg/grpcapi"
	"github.com/alkurbatov/metrics-collector/pkg/metrics"
	"github.com/rs/zerolog/log"
)

func toRecord(ctx context.Context, req *grpcapi.MetricReq, signer *security.Signer) (storage.Record, error) {
	var record storage.Record

	if err := validators.ValidateMetricName(req.Id, req.Mtype); err != nil {
		return record, err
	}

	switch req.Mtype {
	case metrics.KindCounter:
		record = storage.Record{Name: req.Id, Value: metrics.Counter(req.Delta)}

	case metrics.KindGauge:
		record = storage.Record{Name: req.Id, Value: metrics.Gauge(req.Value)}

	default:
		return record, entity.MetricNotImplementedError(req.Mtype)
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

func toMetricReq(record storage.Record, signer *security.Signer) (*grpcapi.MetricReq, error) {
	req := &grpcapi.MetricReq{
		Id:    record.Name,
		Mtype: record.Value.Kind(),
	}

	switch record.Value.Kind() {
	case metrics.KindCounter:
		delta, _ := record.Value.(metrics.Counter)
		req.Delta = int64(delta)

	case metrics.KindGauge:
		value, _ := record.Value.(metrics.Gauge)
		req.Value = float64(value)
	}

	if signer != nil {
		hash, err := signer.CalculateRecordSignature(record)
		if err != nil {
			return nil, fmt.Errorf("grpcbackend - toMetricReq - signer.CalculateRecordSignature: %w", err)
		}

		req.Hash = hash
	}

	return req, nil
}

func toRecordsList(
	ctx context.Context,
	req *grpcapi.BatchUpdateRequest,
	signer *security.Signer,
) ([]storage.Record, error) {
	rv := make([]storage.Record, len(req.Data))

	for i := range req.Data {
		record, err := toRecord(ctx, req.Data[i], signer)
		if err != nil {
			return nil, err
		}

		rv[i] = record
	}

	if len(rv) == 0 {
		return nil, entity.ErrIncompleteRequest
	}

	return rv, nil
}
