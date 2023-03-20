// Package metrics provides client REST API for metrics collector (server).
package metrics

// MetricReq represents info regarding particular metric name, type and value.
// Used in REST API requests/responses to/from metrics collector.
type MetricReq struct {
	// Name of a metric.
	ID string `json:"id"`

	// One of supported metric kinds (e.g. counter, gauge), see constants.
	MType string `json:"type"`

	// Metric value if type is counter, must not be set for other types.
	Delta *Counter `json:"delta,omitempty"`

	// Metric value if type is gauge, must not be set for other types.
	Value *Gauge `json:"value,omitempty"`

	// Hash value of the data, may be omitted if signature validation is
	// disabled on server-side.
	Hash string `json:"hash,omitempty"`
}

// NewUpdateCounterReq creates new MetricReq structure to be used for
// updating counter metric.
func NewUpdateCounterReq(name string, value Counter) MetricReq {
	return MetricReq{ID: name, MType: value.Kind(), Delta: &value}
}

// NewUpdateGaugeReq creates new MetricReq structure to be used for
// updating gauge metric.
func NewUpdateGaugeReq(name string, value Gauge) MetricReq {
	return MetricReq{ID: name, MType: value.Kind(), Value: &value}
}

// NewGetCounterReq creates new MetricReq structure to be used for
// retrieving of counter metric.
func NewGetCounterReq(name string) MetricReq {
	return MetricReq{ID: name, MType: KindCounter}
}

// NewGetGaugeReq creates new MetricReq structure to be used for
// retrieving of gauge metric.
func NewGetGaugeReq(name string) MetricReq {
	return MetricReq{ID: name, MType: KindGauge}
}
