package metrics_test

import (
	"testing"

	"github.com/alkurbatov/metrics-collector/pkg/metrics"
	"github.com/stretchr/testify/assert"
)

func TestStringConvertion(t *testing.T) {
	tt := []struct {
		name     string
		metric   metrics.Metric
		expected string
	}{
		{
			name:     "Convert counter",
			metric:   metrics.Counter(15),
			expected: "15",
		},
		{
			name:     "Convert gauge",
			metric:   metrics.Gauge(15.546789),
			expected: "15.546789",
		},
		{
			name:     "Convert small gauge",
			metric:   metrics.Gauge(0.12),
			expected: "0.12",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.metric.String())
		})
	}
}

func TestKind(t *testing.T) {
	tt := []struct {
		name     string
		metric   metrics.Metric
		expected string
	}{
		{
			name:     "metrics.Counter kind",
			metric:   metrics.Counter(15),
			expected: metrics.KindCounter,
		},
		{
			name:     "metrics.Gauge kind",
			metric:   metrics.Gauge(0.5),
			expected: metrics.KindGauge,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.metric.Kind())
		})
	}
}
