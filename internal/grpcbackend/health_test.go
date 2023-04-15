package grpcbackend_test

import (
	"context"
	"net"
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/grpcbackend"
	"github.com/alkurbatov/metrics-collector/internal/services"
	"github.com/alkurbatov/metrics-collector/pkg/grpcapi"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
)

func sendTestRequest(t *testing.T, healthcheck *services.HealthCheckMock) *status.Status {
	t.Helper()
	require := require.New(t)

	lis := bufconn.Listen(1024 * 1024)
	defer func() {
		require.NoError(lis.Close())
	}()

	srv := grpc.NewServer()
	defer srv.Stop()

	grpcbackend.NewHealthServer(srv, healthcheck)

	go func() {
		require.NoError(srv.Serve(lis))
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial() //nolint: wrapcheck
	}
	conn, err := grpc.Dial("", grpc.WithContextDialer(dialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(err)

	defer func() {
		require.NoError(conn.Close())
	}()

	client := grpcapi.NewHealthClient(conn)
	_, err = client.Ping(context.Background(), new(emptypb.Empty))

	status, ok := status.FromError(err)
	require.True(ok)

	return status
}

func TestPing(t *testing.T) {
	tt := []struct {
		name      string
		checkResp error
		expected  codes.Code
	}{
		{
			name:      "Should return OK, if storage online",
			checkResp: nil,
			expected:  codes.OK,
		},
		{
			name:      "Should return not implemented, if storage doesn't support health check",
			checkResp: entity.ErrHealthCheckNotSupported,
			expected:  codes.Unimplemented,
		},
		{
			name:      "Should return internal error, if storage offline",
			checkResp: entity.ErrUnexpected,
			expected:  codes.Internal,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			m := new(services.HealthCheckMock)
			m.On("CheckStorage", mock.Anything).Return(tc.checkResp)

			status := sendTestRequest(t, m)
			require.Equal(t, tc.expected, status.Code())
		})
	}
}
