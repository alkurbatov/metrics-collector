package services_test

import (
	"context"
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/alkurbatov/metrics-collector/internal/services"
	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pushCounter(
	t *testing.T,
	recorder services.Recorder,
	name string,
	value metrics.Counter,
	expected metrics.Counter,
) {
	t.Helper()

	ctx := context.Background()
	require := require.New(t)

	rv, err := recorder.PushCounter(ctx, name, value)
	require.NoError(err)
	require.Equal(expected, rv)
}

func pushGauge(t *testing.T, recorder services.Recorder, name string, value metrics.Gauge, expected metrics.Gauge) {
	t.Helper()

	ctx := context.Background()
	require := require.New(t)

	rv, err := recorder.PushGauge(ctx, name, value)
	require.NoError(err)
	require.Equal(expected, rv)
}

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
	storage := storage.NewMemStorage()

	for _, tc := range tt {
		ctx := context.Background()
		r := services.NewMetricsRecorder(storage)

		pushCounter(t, r, "PollCount", tc.value, tc.expected)

		record, err := r.GetRecord(ctx, entity.Counter, "PollCount")
		require.NoError(err)
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
	storage := storage.NewMemStorage()

	for _, tc := range tt {
		ctx := context.Background()
		r := services.NewMetricsRecorder(storage)

		pushGauge(t, r, "Alloc", tc.value, tc.expected)

		record, err := r.GetRecord(ctx, entity.Gauge, "Alloc")
		require.NoError(err)
		require.Equal(tc.expected, record.Value)
	}
}

func TestPushMetricsWithSimilarNamesButDifferentKinds(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()
	r := services.NewMetricsRecorder(storage.NewMemStorage())

	pushCounter(t, r, "X", 10, 10)
	pushGauge(t, r, "X", 20.123, 20.123)

	first, err := r.GetRecord(ctx, entity.Counter, "X")
	require.NoError(err)
	require.Equal(metrics.Counter(10), first.Value)

	second, err := r.GetRecord(ctx, entity.Gauge, "X")
	require.NoError(err)
	require.Equal(metrics.Gauge(20.123), second.Value)
}

func TestPushMetricsToBrokenStorage(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()
	store := &storage.BrokenStorage{}
	r := services.NewMetricsRecorder(store)

	_, err := r.PushCounter(ctx, "PollCount", metrics.Counter(1))
	require.Error(err)

	_, err = r.PushGauge(ctx, "Alloc", metrics.Gauge(13.2))
	require.Error(err)
}

func TestGetUnknownMetric(t *testing.T) {
	tt := []struct {
		name     string
		kind     string
		metric   string
		expected error
	}{
		{
			name:     "Should return error on unknown counter",
			kind:     entity.Counter,
			metric:   "unknown",
			expected: entity.ErrMetricNotFound,
		},
		{
			name:     "Should return error on unknown gauge",
			kind:     entity.Gauge,
			metric:   "unknown",
			expected: entity.ErrMetricNotFound,
		},
		{
			name:     "Should return error on unknown kind",
			kind:     "unknown",
			metric:   "PollCounter",
			expected: entity.ErrMetricNotImplemented,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			r := services.RecorderMock{}

			_, err := r.GetRecord(ctx, tc.kind, tc.metric)
			assert.ErrorIs(t, err, tc.expected)
		})
	}
}

func TestListMetrics(t *testing.T) {
	require := require.New(t)
	r := services.NewMetricsRecorder(storage.NewMemStorage())

	pushCounter(t, r, "PollCount", 10, 10)
	pushGauge(t, r, "Alloc", 11.123, 11.123)

	expected := []storage.Record{
		{Name: "Alloc", Value: metrics.Gauge(11.123)},
		{Name: "PollCount", Value: metrics.Counter(10)},
	}

	data, err := r.ListRecords(context.Background())

	require.NoError(err)
	require.Equal(2, len(data))
	require.Equal(expected, data)
}
