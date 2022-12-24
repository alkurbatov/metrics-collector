package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringConvertion(t *testing.T) {
	tt := []struct {
		name     string
		metric   Metric
		expected string
	}{
		{
			name:     "Convert counter",
			metric:   Counter(15),
			expected: "15",
		},
		{
			name:     "Convert gauge",
			metric:   Gauge(15.546789),
			expected: "15.546789",
		},
		{
			name:     "Convert small gauge",
			metric:   Gauge(0.12),
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
		metric   Metric
		expected string
	}{
		{
			name:     "Counter kind",
			metric:   Counter(15),
			expected: "counter",
		},
		{
			name:     "Gauge kind",
			metric:   Gauge(0.5),
			expected: "gauge",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.metric.Kind())
		})
	}
}
