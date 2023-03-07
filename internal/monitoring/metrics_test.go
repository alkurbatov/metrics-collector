package monitoring_test

import (
	"context"
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/monitoring"
	"github.com/alkurbatov/metrics-collector/pkg/metrics"
	"github.com/stretchr/testify/require"
)

func TestPoll(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	m := new(monitoring.Metrics)
	require.Zero(m.PollCount)
	require.Zero(m.RandomValue)
	require.Zero(m.System.TotalMemory)
	require.Zero(m.Runtime.Alloc)

	err := m.Poll(ctx)

	require.NoError(err)
	require.Equal(m.PollCount, metrics.Counter(1))
	require.NotZero(m.RandomValue)
	require.NotZero(m.System.TotalMemory)
	require.NotZero(m.Runtime.Alloc)

	old := *m
	err = m.Poll(ctx)

	require.NoError(err)
	require.Equal(m.PollCount, metrics.Counter(2))
	require.NotEqual(old.RandomValue, m.RandomValue)
	require.NotEqual(old.Runtime.Alloc, m.Runtime.Alloc)
}
