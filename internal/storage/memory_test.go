package storage_test

import (
	"context"
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const metricName = "PollCount"
const metricID = "PollCount_counter"

func TestPush(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()
	m := storage.NewMemStorage()

	value := metrics.Counter(10)
	err := m.Push(ctx, metricID, storage.Record{Name: metricName, Value: value})
	require.NoError(err)
	require.Equal(value, m.Data[metricID].Value)

	value = metrics.Counter(23)
	err = m.Push(ctx, metricID, storage.Record{Name: metricName, Value: value})
	require.NoError(err)
	require.Equal(value, m.Data[metricID].Value)
}

func TestPushList(t *testing.T) {
	keys := []string{
		"PollCount_counter",
		"Alloc_gauge",
	}
	input := []storage.Record{
		{Name: "PollCount", Value: metrics.Counter(10)},
		{Name: "Alloc", Value: metrics.Gauge(13.123)},
	}

	tt := []struct {
		name    string
		keys    []string
		records []storage.Record
	}{
		{
			name:    "Should push list of records",
			keys:    keys,
			records: input,
		},
		{
			name:    "Should not fail on empty input",
			keys:    make([]string, 0),
			records: make([]storage.Record, 0),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			m := storage.NewMemStorage()

			err := m.PushList(context.Background(), tc.keys, tc.records)
			assert.NoError(t, err)
		})
	}
}

func TestGet(t *testing.T) {
	ctx := context.Background()
	value := metrics.Counter(10)

	require := require.New(t)
	m := storage.NewMemStorage()
	err := m.Push(ctx, metricID, storage.Record{Name: metricName, Value: value})
	require.NoError(err)

	record, err := m.Get(ctx, metricID)
	require.NoError(err)
	require.Equal(value, record.Value)
}

func TestGetUnknownstorageRecord(t *testing.T) {
	m := storage.NewMemStorage()

	_, err := m.Get(context.Background(), "XXX")
	require.ErrorIs(t, err, entity.ErrMetricNotFound)
}

func TestGetAll(t *testing.T) {
	ctx := context.Background()
	require := require.New(t)
	keys := []string{"Alloc_gauge", "PollCount_counter", "Random_gauge"}
	input := []storage.Record{
		{Name: "Alloc", Value: metrics.Gauge(11.123)},
		{Name: "PollCount", Value: metrics.Counter(10)},
		{Name: "Random", Value: metrics.Gauge(33.3333)},
	}

	m := storage.NewMemStorage()
	for i, key := range keys {
		err := m.Push(ctx, key, input[i])
		require.NoError(err)
	}

	records, err := m.GetAll(ctx)
	require.NoError(err)
	require.ElementsMatch(input, records)

	err = m.Push(ctx, "New_counter", storage.Record{Name: "New", Value: metrics.Counter(1)})
	require.NoError(err)
	require.Equal(len(input), len(records))
}

func TestGetAllOnEmptyStorage(t *testing.T) {
	m := storage.NewMemStorage()
	records, err := m.GetAll(context.Background())

	assert.NoError(t, err)
	assert.Empty(t, records)
}

func TestSnapshot(t *testing.T) {
	id := "PollCount_counter"
	name := "PollCount"

	require := require.New(t)
	ctx := context.Background()
	m := storage.NewMemStorage()

	value := metrics.Counter(10)
	err := m.Push(ctx, id, storage.Record{Name: name, Value: value})
	require.NoError(err)
	require.Equal(value, m.Data[id].Value)

	snapshot := m.Snapshot()
	snapshotMetrics, err := m.Snapshot().GetAll(ctx)
	require.NoError(err)

	currentMetrics, err := m.GetAll(ctx)
	require.NoError(err)
	require.ElementsMatch(currentMetrics, snapshotMetrics)

	newValue := metrics.Counter(22)
	err = m.Push(ctx, id, storage.Record{Name: name, Value: newValue})
	require.NoError(err)
	require.Equal(newValue, m.Data[id].Value)
	require.Equal(value, snapshot.Data[id].Value)
}

func TestCloseIsNoop(t *testing.T) {
	m := storage.NewMemStorage()
	assert.NoError(t, m.Close())
}
