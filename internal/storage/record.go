package storage

import (
	"encoding/json"
	"fmt"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/pkg/metrics"
)

func unmarshalError(reason error) error {
	return fmt.Errorf("record unmarshaling failed: %w", reason)
}

type Record struct {
	Name  string
	Value metrics.Metric
}

func (r Record) MarshalJSON() ([]byte, error) {
	rv, err := json.Marshal(map[string]string{
		"name":  r.Name,
		"kind":  r.Value.Kind(),
		"value": r.Value.String(),
	})

	if err != nil {
		return nil, fmt.Errorf("record marshaling failed: %w", err)
	}

	return rv, nil
}

func (r *Record) UnmarshalJSON(src []byte) error {
	var data map[string]string
	if err := json.Unmarshal(src, &data); err != nil {
		return unmarshalError(err)
	}

	r.Name = data["name"]

	switch data["kind"] {
	case metrics.KindCounter:
		value, err := metrics.ToCounter(data["value"])
		if err != nil {
			return unmarshalError(err)
		}

		r.Value = value

	case metrics.KindGauge:
		value, err := metrics.ToGauge(data["value"])
		if err != nil {
			return unmarshalError(err)
		}

		r.Value = value

	default:
		return unmarshalError(entity.MetricNotImplementedError(data["kind"]))
	}

	return nil
}
