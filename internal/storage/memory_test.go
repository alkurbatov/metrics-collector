package storage

import (
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/stretchr/testify/require"
)

func TestPush(t *testing.T) {
	name := "PollCount_counter"

	m := NewMemStorage()

	value := metrics.Counter(10)
	m.Push(name, Record{Name: name, Value: value})
	require.Equal(t, value, m.data[name].Value)

	value = metrics.Counter(23)
	m.Push(name, Record{Name: name, Value: value})
	require.Equal(t, value, m.data[name].Value)
}

func TestGet(t *testing.T) {
	key := "PollCount_counter"
	name := "PollCounter"
	value := metrics.Counter(10)

	m := NewMemStorage()
	m.Push(key, Record{Name: name, Value: value})

	record, ok := m.Get(key)
	require.True(t, ok)
	require.Equal(t, value, record.Value)
}

func TestGetUnknownRecord(t *testing.T) {
	m := NewMemStorage()

	_, ok := m.Get("XXX")
	require.False(t, ok)
}

func TestGetAll(t *testing.T) {
	require := require.New(t)
	keys := []string{"Alloc_gauge", "PollCount_counter", "Random_gauge"}
	input := []Record{
		{Name: "Alloc", Value: metrics.Gauge(11.123)},
		{Name: "PollCount", Value: metrics.Counter(10)},
		{Name: "Random", Value: metrics.Gauge(33.3333)},
	}

	m := NewMemStorage()
	for i, key := range keys {
		m.Push(key, input[i])
	}

	records := m.GetAll()
	require.ElementsMatch(input, records)

	m.Push("New_counter", Record{Name: "New", Value: metrics.Counter(1)})
	require.Equal(len(input), len(records))
}
