package schema

import (
	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/metrics"
)

type MetricReq struct {
	ID    string           `json:"id"`              // name of a metric
	MType string           `json:"type"`            // gauge or counter
	Delta *metrics.Counter `json:"delta,omitempty"` // metric value if type is counter
	Value *metrics.Gauge   `json:"value,omitempty"` // metric value if type is gauge
	Hash  string           `json:"hash,omitempty"`  // hash value of the data
}

func NewUpdateCounterReq(name string, value metrics.Counter) MetricReq {
	return MetricReq{ID: name, MType: value.Kind(), Delta: &value}
}

func NewUpdateGaugeReq(name string, value metrics.Gauge) MetricReq {
	return MetricReq{ID: name, MType: value.Kind(), Value: &value}
}

func NewGetCounterReq(name string) MetricReq {
	return MetricReq{ID: name, MType: entity.Counter}
}

func NewGetGaugeReq(name string) MetricReq {
	return MetricReq{ID: name, MType: entity.Gauge}
}
