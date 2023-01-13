package schema_test

import (
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/alkurbatov/metrics-collector/internal/schema"
	"github.com/stretchr/testify/require"
)

func TestNewMetricReq(t *testing.T) {
	require := require.New(t)

	delta := metrics.Counter(1)
	expected := schema.MetricReq{ID: "xxx", MType: entity.Counter, Delta: &delta}
	require.Equal(expected, schema.NewUpdateCounterReq("xxx", 1))

	value := metrics.Gauge(1.12)
	expected = schema.MetricReq{ID: "xxx", MType: entity.Gauge, Value: &value}
	require.Equal(expected, schema.NewUpdateGaugeReq("xxx", 1.12))

	expected = schema.MetricReq{ID: "xxx", MType: entity.Counter}
	require.Equal(expected, schema.NewGetCounterReq("xxx"))

	expected = schema.MetricReq{ID: "xxx", MType: entity.Gauge}
	require.Equal(expected, schema.NewGetGaugeReq("xxx"))
}
