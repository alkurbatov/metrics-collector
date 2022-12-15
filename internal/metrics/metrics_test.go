package metrics

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPoll(t *testing.T) {
	require := require.New(t)

	m := new(Metrics)
	require.Zero(m.PollCount)
	require.Zero(m.RandomValue)

	m.Poll()
	require.Equal(m.PollCount, Counter(1))
	require.NotZero(m.RandomValue)

	oldGauge := m.RandomValue
	m.Poll()
	require.Equal(m.PollCount, Counter(2))
	require.NotEqual(oldGauge, m.RandomValue)
}
