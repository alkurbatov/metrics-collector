package services_test

import (
	"context"
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/alkurbatov/metrics-collector/internal/services"
	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func pushMetric(
	t *testing.T,
	recorder services.Recorder,
	name string,
	value metrics.Metric,
	expected metrics.Metric,
) {
	t.Helper()

	ctx := context.Background()
	require := require.New(t)

	rv, err := recorder.Push(ctx, storage.Record{Name: name, Value: value})
	require.NoError(err)
	require.Equal(storage.Record{Name: name, Value: expected}, rv)
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

		pushMetric(t, r, "PollCount", tc.value, tc.expected)

		record, err := r.Get(ctx, entity.Counter, "PollCount")
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

		pushMetric(t, r, "Alloc", tc.value, tc.expected)

		record, err := r.Get(ctx, entity.Gauge, "Alloc")
		require.NoError(err)
		require.Equal(tc.expected, record.Value)
	}
}

func TestPushMetricsWithSimilarNamesButDifferentKinds(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()
	r := services.NewMetricsRecorder(storage.NewMemStorage())

	delta := metrics.Counter(10)
	pushMetric(t, r, "X", delta, delta)

	value := metrics.Gauge(20.123)
	pushMetric(t, r, "X", value, value)

	first, err := r.Get(ctx, entity.Counter, "X")
	require.NoError(err)
	require.Equal(metrics.Counter(10), first.Value)

	second, err := r.Get(ctx, entity.Gauge, "X")
	require.NoError(err)
	require.Equal(metrics.Gauge(20.123), second.Value)
}

func TestPushMetricsToBrokenStorage(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	store := new(storage.Mock)
	store.On("Get", ctx, "NotFound_counter").Return(storage.Record{}, entity.ErrMetricNotFound)
	store.On("Get", ctx, mock.AnythingOfType("string")).Return(storage.Record{}, entity.ErrUnexpected)
	store.On("Push", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("Record")).Return(entity.ErrUnexpected)

	r := services.NewMetricsRecorder(store)

	_, err := r.Push(ctx, storage.Record{Name: "PollCount", Value: metrics.Counter(1)})
	require.Error(err)

	_, err = r.Push(ctx, storage.Record{Name: "NotFound", Value: metrics.Counter(1)})
	require.Error(err)

	_, err = r.Push(ctx, storage.Record{Name: "Alloc", Value: metrics.Gauge(13.2)})
	require.Error(err)

	store.AssertExpectations(t)
}

func TestPushListTestPushList(t *testing.T) {
	type expected struct {
		keys    []string
		records []storage.Record
		err     error
	}

	tt := []struct {
		name       string
		records    []storage.Record
		storageErr error
		expected   expected
	}{
		{
			name: "Should push metrics",
			records: []storage.Record{
				{Name: "PollCount", Value: metrics.Counter(10)},
				{Name: "Alloc", Value: metrics.Gauge(10.123)},
			},
			expected: expected{
				keys: []string{"PollCount_counter", "Alloc_gauge"},
				records: []storage.Record{
					{Name: "PollCount", Value: metrics.Counter(11)},
					{Name: "Alloc", Value: metrics.Gauge(10.123)},
				},
			},
		},
		{
			name: "Should compress duplicated metrics",
			records: []storage.Record{
				{Name: "PollCount", Value: metrics.Counter(10)},
				{Name: "Alloc", Value: metrics.Gauge(10.123)},
				{Name: "PollCount", Value: metrics.Counter(12)},
				{Name: "Alloc", Value: metrics.Gauge(14.321)},
			},
			expected: expected{
				keys: []string{"PollCount_counter", "Alloc_gauge"},
				records: []storage.Record{
					{Name: "PollCount", Value: metrics.Counter(23)},
					{Name: "Alloc", Value: metrics.Gauge(14.321)},
				},
			},
		},
		{
			name:    "Should not fail on empty list",
			records: make([]storage.Record, 0),
			expected: expected{
				keys:    make([]string, 0),
				records: make([]storage.Record, 0),
			},
		},
		{
			name: "Should fail on broken storage",
			records: []storage.Record{
				{Name: "PollCount", Value: metrics.Counter(10)},
			},
			storageErr: entity.ErrUnexpected,
			expected: expected{
				keys:    make([]string, 0),
				records: make([]storage.Record, 0),
				err:     entity.ErrUnexpected,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			m := new(storage.Mock)
			m.On("Get", mock.Anything, "PollCount_counter").
				Return(storage.Record{Name: "PollCount", Value: metrics.Counter(1)}, tc.storageErr)
			m.On("Get", mock.Anything, mock.AnythingOfType("string")).
				Return(storage.Record{}, entity.ErrMetricNotFound)
			m.On("PushList", mock.Anything, tc.expected.keys, tc.expected.records).
				Return(tc.storageErr)

			r := services.NewMetricsRecorder(m)
			err := r.PushList(context.Background(), tc.records)

			assert.ErrorIs(t, err, tc.expected.err)
		})
	}
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
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			store := new(storage.Mock)
			store.On("Get", ctx, mock.AnythingOfType("string")).Return(storage.Record{}, entity.ErrMetricNotFound)
			r := services.NewMetricsRecorder(store)

			_, err := r.Get(ctx, tc.kind, tc.metric)
			assert.ErrorIs(t, err, tc.expected)
			store.AssertExpectations(t)
		})
	}
}

func TestListMetrics(t *testing.T) {
	stored := []storage.Record{
		{Name: "PollCount", Value: metrics.Counter(10)},
		{Name: "Alloc", Value: metrics.Gauge(11.123)},
	}

	expected := []storage.Record{
		{Name: "Alloc", Value: metrics.Gauge(11.123)},
		{Name: "PollCount", Value: metrics.Counter(10)},
	}

	m := new(storage.Mock)
	m.On("GetAll", mock.Anything).Return(stored, nil)

	require := require.New(t)
	r := services.NewMetricsRecorder(m)

	data, err := r.List(context.Background())

	require.NoError(err)
	require.Equal(2, len(data))
	require.Equal(expected, data)
}

func TestListMetricsOnBrokenStorage(t *testing.T) {
	store := new(storage.Mock)
	store.On("GetAll", mock.Anything).Return(nil, entity.ErrUnexpected)

	r := services.NewMetricsRecorder(store)
	_, err := r.List(context.Background())

	require.ErrorIs(t, entity.ErrUnexpected, err)

	store.AssertExpectations(t)
}
