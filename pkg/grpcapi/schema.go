// Package grpcapi provides protobuf bindings to use with gRPC
// and various helper functions.
package grpcapi

import "github.com/alkurbatov/metrics-collector/pkg/metrics"

// NewUpdateCounterReq creates new MetricReq structure to be used for
// updating counter metric.
func NewUpdateCounterReq(name string, value metrics.Counter) *MetricReq {
	return &MetricReq{Id: name, Mtype: value.Kind(), Delta: int64(value)}
}

// NewUpdateGaugeReq creates new MetricReq structure to be used for
// updating gauge metric.
func NewUpdateGaugeReq(name string, value metrics.Gauge) *MetricReq {
	return &MetricReq{Id: name, Mtype: value.Kind(), Value: float64(value)}
}

// NewGetCounterReq creates new GetMetricRequest structure to be used for
// retrieving of counter metric.
func NewGetCounterReq(name string) *GetMetricRequest {
	return &GetMetricRequest{Id: name, Mtype: metrics.KindCounter}
}

// NewGetGaugeReq creates new GetMetricRequest structure to be used for
// retrieving of gauge metric.
func NewGetGaugeReq(name string) *GetMetricRequest {
	return &GetMetricRequest{Id: name, Mtype: metrics.KindGauge}
}
