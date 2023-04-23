package grpcbackend_test

import (
	"context"
	"net"
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/grpcbackend"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/alkurbatov/metrics-collector/internal/services"
	"github.com/alkurbatov/metrics-collector/pkg/grpcapi"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func requireEqual(t *testing.T, left *grpcapi.MetricReq, right *grpcapi.MetricReq) {
	t.Helper()
	require := require.New(t)

	require.Equal(left.Id, right.Id)
	require.Equal(left.Mtype, right.Mtype)
	require.Equal(left.Delta, right.Delta)
	require.Equal(left.Value, right.Value)
	require.Equal(left.Hash, right.Hash)
}

func requireEqualCode(t *testing.T, expected codes.Code, err error) {
	t.Helper()

	rv, ok := status.FromError(err)

	require.True(t, ok)
	require.Equal(t, expected, rv.Code())
}

func createTestServer(
	t *testing.T,
	recorder *services.RecorderMock,
	healthcheck *services.HealthCheckMock,
	key security.Secret,
) (*grpc.ClientConn, func()) {
	t.Helper()
	require := require.New(t)

	if recorder == nil {
		recorder = &services.RecorderMock{}
	}

	if healthcheck == nil {
		healthcheck = &services.HealthCheckMock{}
	}

	var signer *security.Signer
	if len(key) != 0 {
		signer = security.NewSigner(key)
	}

	lis := bufconn.Listen(1024 * 1024)
	srv := grpcbackend.New("", recorder, healthcheck, signer, nil).Instance()

	go func() {
		require.NoError(srv.Serve(lis))
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
	conn, err := grpc.Dial("", grpc.WithContextDialer(dialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(err)

	closer := func() {
		require.NoError(conn.Close())
		srv.Stop()
		require.NoError(lis.Close())
	}

	return conn, closer
}
