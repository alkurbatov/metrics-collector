package storage_test

import (
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const metricName = "PollCount"
const metricID = "PollCount_counter"

func TestPush(t *testing.T) {
	require := require.New(t)
	m := storage.NewMemStorage()

	value := metrics.Counter(10)
	err := m.Push(metricID, storage.Record{Name: metricName, Value: value})
	require.NoError(err)
	require.Equal(value, m.Data[metricID].Value)

	value = metrics.Counter(23)
	err = m.Push(metricID, storage.Record{Name: metricName, Value: value})
	require.NoError(err)
	require.Equal(value, m.Data[metricID].Value)
}

func TestGet(t *testing.T) {
	value := metrics.Counter(10)

	require := require.New(t)
	m := storage.NewMemStorage()
	err := m.Push(metricID, storage.Record{Name: metricName, Value: value})
	require.NoError(err)

	record, ok := m.Get(metricID)
	require.True(ok)
	require.Equal(value, record.Value)
}

func TestGetUnknownstorageRecord(t *testing.T) {
	m := storage.NewMemStorage()

	_, ok := m.Get("XXX")
	require.False(t, ok)
}

func TestGetAll(t *testing.T) {
	require := require.New(t)
	keys := []string{"Alloc_gauge", "PollCount_counter", "Random_gauge"}
	input := []storage.Record{
		{Name: "Alloc", Value: metrics.Gauge(11.123)},
		{Name: "PollCount", Value: metrics.Counter(10)},
		{Name: "Random", Value: metrics.Gauge(33.3333)},
	}

	m := storage.NewMemStorage()
	for i, key := range keys {
		err := m.Push(key, input[i])
		require.NoError(err)
	}

	records := m.GetAll()
	require.ElementsMatch(input, records)

	err := m.Push("New_counter", storage.Record{Name: "New", Value: metrics.Counter(1)})
	require.NoError(err)
	require.Equal(len(input), len(records))
}

func TestGetAllOnEmptyStorage(t *testing.T) {
	m := storage.NewMemStorage()
	records := m.GetAll()

	assert.Empty(t, records)
}

func TestSnapshot(t *testing.T) {
	id := "PollCount_counter"
	name := "PollCount"

	require := require.New(t)
	m := storage.NewMemStorage()

	value := metrics.Counter(10)
	err := m.Push(id, storage.Record{Name: name, Value: value})
	require.NoError(err)
	require.Equal(value, m.Data[id].Value)

	snapshot := m.Snapshot()
	require.ElementsMatch(m.GetAll(), snapshot.GetAll())

	newValue := metrics.Counter(22)
	err = m.Push(id, storage.Record{Name: name, Value: newValue})
	require.NoError(err)
	require.Equal(newValue, m.Data[id].Value)
	require.Equal(value, snapshot.Data[id].Value)
}
