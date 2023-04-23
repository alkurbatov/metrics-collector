package grpcapi_test

import (
	"testing"

	"github.com/alkurbatov/metrics-collector/pkg/grpcapi"
	"github.com/alkurbatov/metrics-collector/pkg/metrics"
	"github.com/stretchr/testify/require"
)

func TestNewUpdateMetricReq(t *testing.T) {
	require := require.New(t)

	delta := metrics.Counter(1)
	expected := &grpcapi.MetricReq{Id: "xxx", Mtype: metrics.KindCounter, Delta: int64(delta)}
	require.Equal(expected, grpcapi.NewUpdateCounterReq("xxx", 1))

	value := metrics.Gauge(1.12)
	expected = &grpcapi.MetricReq{Id: "xxx", Mtype: metrics.KindGauge, Value: float64(value)}
	require.Equal(expected, grpcapi.NewUpdateGaugeReq("xxx", 1.12))
}

func TestNewGetMetricReq(t *testing.T) {
	require := require.New(t)

	expected := &grpcapi.GetMetricRequest{Id: "xxx", Mtype: metrics.KindCounter}
	require.Equal(expected, grpcapi.NewGetCounterReq("xxx"))

	expected = &grpcapi.GetMetricRequest{Id: "xxx", Mtype: metrics.KindGauge}
	require.Equal(expected, grpcapi.NewGetGaugeReq("xxx"))
}
