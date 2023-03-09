package metrics_test

import (
	"testing"

	"github.com/alkurbatov/metrics-collector/pkg/metrics"
	"github.com/stretchr/testify/require"
)

func TestNewMetricReq(t *testing.T) {
	require := require.New(t)

	delta := metrics.Counter(1)
	expected := metrics.MetricReq{ID: "xxx", MType: metrics.KindCounter, Delta: &delta}
	require.Equal(expected, metrics.NewUpdateCounterReq("xxx", 1))

	value := metrics.Gauge(1.12)
	expected = metrics.MetricReq{ID: "xxx", MType: metrics.KindGauge, Value: &value}
	require.Equal(expected, metrics.NewUpdateGaugeReq("xxx", 1.12))

	expected = metrics.MetricReq{ID: "xxx", MType: metrics.KindCounter}
	require.Equal(expected, metrics.NewGetCounterReq("xxx"))

	expected = metrics.MetricReq{ID: "xxx", MType: metrics.KindGauge}
	require.Equal(expected, metrics.NewGetGaugeReq("xxx"))
}
