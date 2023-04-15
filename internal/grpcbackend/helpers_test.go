package grpcbackend_test

import (
	"testing"

	"github.com/alkurbatov/metrics-collector/pkg/grpcapi"
	"github.com/stretchr/testify/require"
)

func requireEqual(t *testing.T, left *grpcapi.MetricReq, right *grpcapi.MetricReq) {
	t.Helper()
	require := require.New(t)

	require.Equal(left.Id, right.Id)
	require.Equal(left.Mtype, right.Mtype)
	require.Equal(left.Delta, right.Delta)
	require.Equal(left.Value, right.Value)
}
