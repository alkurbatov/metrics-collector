package grpcbackend_test

import (
	"context"
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/services"
	"github.com/alkurbatov/metrics-collector/pkg/grpcapi"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
)

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

			conn, closer := createTestServer(t, nil, m, "")
			defer closer()

			client := grpcapi.NewHealthClient(conn)
			_, err := client.Ping(context.Background(), new(emptypb.Empty))

			requireEqualCode(t, tc.expected, err)
		})
	}
}
