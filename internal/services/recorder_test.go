package services

import (
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/app"
	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateCounter(t *testing.T) {
	tt := []struct {
		value    metrics.Counter
		expected metrics.Counter
	}{
		{
			value:    3,
			expected: 3,
		},
		{
			value:    5,
			expected: 8,
		},
		{
			value:    13,
			expected: 21,
		},
	}

	require := require.New(t)
	app := app.NewServer()

	for _, tc := range tt {
		r := NewMetricsRecorder(app)

		r.PushCounter("PollCount", tc.value)

		record, ok := r.GetRecord("counter", "PollCount")
		require.True(ok)
		require.Equal(tc.expected, record.Value)
	}
}

func TestUpdateGauge(t *testing.T) {
	tt := []struct {
		value    metrics.Gauge
		expected metrics.Gauge
	}{
		{
			value:    3.123000,
			expected: 3.123,
		},
		{
			value:    5.456230,
			expected: 5.45623,
		},
		{
			value:    13.123856,
			expected: 13.123856,
		},
	}

	require := require.New(t)
	app := app.NewServer()

	for _, tc := range tt {
		r := NewMetricsRecorder(app)

		r.PushGauge("Alloc", tc.value)

		record, ok := r.GetRecord("gauge", "Alloc")
		require.True(ok)
		require.Equal(tc.expected, record.Value)
	}
}

func TestPushMetricsWithSimilarNamesButDifferentKinds(t *testing.T) {
	require := require.New(t)
	app := app.NewServer()
	r := NewMetricsRecorder(app)

	r.PushCounter("X", 10)
	r.PushGauge("X", 20.123)

	first, ok := r.GetRecord("counter", "X")
	require.True(ok)
	require.Equal(metrics.Counter(10), first.Value)

	second, ok := r.GetRecord("gauge", "X")
	require.True(ok)
	require.Equal(metrics.Gauge(20.123), second.Value)
}

func TestGetUnknownMetric(t *testing.T) {
	tt := []struct {
		name   string
		kind   string
		metric string
	}{
		{
			name:   "Unknown counter",
			kind:   "counter",
			metric: "XXX",
		},
		{
			name:   "unknown gauge",
			kind:   "gauge",
			metric: "XXX",
		},
		{
			name:   "Unknown kind",
			kind:   "unknown",
			metric: "PollCounter",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			app := app.NewServer()
			r := NewMetricsRecorder(app)

			_, ok := r.GetRecord(tc.kind, tc.metric)
			assert.False(t, ok)
		})
	}
}

func TestListMetrics(t *testing.T) {
	require := require.New(t)
	app := app.NewServer()
	r := NewMetricsRecorder(app)

	r.PushCounter("PollCount", 10)
	r.PushGauge("Alloc", 11.123)
	expected := []storage.Record{
		{Name: "Alloc", Value: metrics.Gauge(11.123)},
		{Name: "PollCount", Value: metrics.Counter(10)},
	}

	data := r.ListRecords()

	require.Equal(2, len(data))
	require.Equal(expected, data)
}
