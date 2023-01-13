package metrics_test

import (
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/stretchr/testify/require"
)

func TestPoll(t *testing.T) {
	require := require.New(t)

	m := new(metrics.Metrics)
	require.Zero(m.PollCount)
	require.Zero(m.RandomValue)

	m.Poll()
	require.Equal(m.PollCount, metrics.Counter(1))
	require.NotZero(m.RandomValue)

	oldGauge := m.RandomValue
	m.Poll()
	require.Equal(m.PollCount, metrics.Counter(2))
	require.NotEqual(oldGauge, m.RandomValue)
}
