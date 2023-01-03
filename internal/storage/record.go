package storage

import (
	"encoding/json"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/metrics"
)

type Record struct {
	Name  string
	Value metrics.Metric
}

func (r Record) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"name":  r.Name,
		"kind":  r.Value.Kind(),
		"value": r.Value.String(),
	})
}

func (r *Record) UnmarshalJSON(src []byte) error {
	var data map[string]string
	if err := json.Unmarshal(src, &data); err != nil {
		return err
	}

	r.Name = data["name"]

	switch data["kind"] {
	case entity.Counter:
		value, err := metrics.ToCounter(data["value"])
		if err != nil {
			return err
		}

		r.Value = value

	case entity.Gauge:
		value, err := metrics.ToGauge(data["value"])
		if err != nil {
			return err
		}

		r.Value = value

	default:
		return &entity.MetricNotImplementedError{Kind: data["kind"]}
	}

	return nil
}
