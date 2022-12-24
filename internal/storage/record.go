package storage

import (
	"encoding/json"
	"errors"

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
	case "counter":
		value, err := metrics.ToCounter(data["value"])
		if err != nil {
			return err
		}

		r.Value = value

	case "gauge":
		value, err := metrics.ToGauge(data["value"])
		if err != nil {
			return err
		}

		r.Value = value

	default:
		return errors.New("unexpected record kind: " + data["kind"])
	}

	return nil
}
