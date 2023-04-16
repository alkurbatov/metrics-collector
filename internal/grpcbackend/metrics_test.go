package grpcbackend_test

import (
	"context"
	"net"
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/grpcbackend"
	"github.com/alkurbatov/metrics-collector/internal/services"
	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/alkurbatov/metrics-collector/pkg/grpcapi"
	"github.com/alkurbatov/metrics-collector/pkg/metrics"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func createTestServer(t *testing.T, recorder *services.RecorderMock) (grpcapi.MetricsClient, func()) {
	t.Helper()
	require := require.New(t)

	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()

	grpcbackend.NewMetricsServer(srv, recorder)

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

	client := grpcapi.NewMetricsClient(conn)

	return client, closer
}

func TestUpdateMetric(t *testing.T) {
	tt := []struct {
		name        string
		req         *grpcapi.MetricReq
		recorderRV  storage.Record
		recorderErr error
		expected    codes.Code
	}{
		{
			name:       "Should push counter",
			req:        grpcapi.NewUpdateCounterReq("PollCount", 10),
			recorderRV: storage.Record{Name: "PollCount", Value: metrics.Counter(10)},
			expected:   codes.OK,
		},
		{
			name:       "Should push gauge",
			req:        grpcapi.NewUpdateGaugeReq("Alloc", 13.123),
			recorderRV: storage.Record{Name: "Alloc", Value: metrics.Gauge(13.123)},
			expected:   codes.OK,
		},
		{
			name: "Should fail on unknown metric kind",
			req: &grpcapi.MetricReq{
				Id:    "X",
				Mtype: "unknown",
			},
			expected: codes.Unimplemented,
		},
		{
			name:     "Should fail on counter with invalid name",
			req:      grpcapi.NewUpdateCounterReq("X)", 10),
			expected: codes.InvalidArgument,
		},
		{
			name:     "Should fail on gauge with invalid name",
			req:      grpcapi.NewUpdateGaugeReq("X;", 13.123),
			expected: codes.InvalidArgument,
		},
		{
			name:        "Should fail on broken recorder",
			req:         grpcapi.NewUpdateGaugeReq("fail", 13),
			recorderErr: entity.ErrUnexpected,
			expected:    codes.Internal,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			m := new(services.RecorderMock)
			m.On("Push", mock.Anything, mock.AnythingOfType("Record")).Return(tc.recorderRV, tc.recorderErr)

			client, closer := createTestServer(t, m)
			t.Cleanup(closer)

			resp, err := client.Update(context.Background(), tc.req)

			status, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tc.expected, status.Code())

			if tc.expected == codes.OK {
				requireEqual(t, tc.req, resp)
			}
		})
	}
}

func TestGetMetric(t *testing.T) {
	type result struct {
		code codes.Code
		body *grpcapi.MetricReq
	}

	tt := []struct {
		name        string
		req         *grpcapi.GetMetricRequest
		recorderRV  storage.Record
		recorderErr error
		expected    result
	}{
		{
			name:       "Should get counter",
			req:        grpcapi.NewGetCounterReq("PollCount"),
			recorderRV: storage.Record{Name: "PollCount", Value: metrics.Counter(10)},
			expected: result{
				code: codes.OK,
				body: grpcapi.NewUpdateCounterReq("PollCount", 10),
			},
		},
		{
			name:       "Should get gauge",
			req:        grpcapi.NewGetGaugeReq("Alloc"),
			recorderRV: storage.Record{Name: "Alloc", Value: metrics.Gauge(11.345)},
			expected: result{
				code: codes.OK,
				body: grpcapi.NewUpdateGaugeReq("Alloc", 11.345),
			},
		},
		{
			name: "Should fail on unknown metric kind",
			req:  &grpcapi.GetMetricRequest{Id: "Alloc", Mtype: "unknown"},
			expected: result{
				code: codes.Unimplemented,
			},
		},
		{
			name:        "Should fail on unknown counter",
			req:         grpcapi.NewGetCounterReq("unknown"),
			recorderErr: entity.ErrMetricNotFound,
			expected: result{
				code: codes.NotFound,
			},
		},
		{
			name:        "Should fail on unknown gauge",
			req:         grpcapi.NewGetGaugeReq("unknown"),
			recorderErr: entity.ErrMetricNotFound,
			expected: result{
				code: codes.NotFound,
			},
		},
		{
			name: "Should fail on counter with invalid name",
			req:  grpcapi.NewGetCounterReq("X)"),
			expected: result{
				code: codes.InvalidArgument,
			},
		},
		{
			name: "Should fail on gauge with invalid name",
			req:  grpcapi.NewGetGaugeReq("X;"),
			expected: result{
				code: codes.InvalidArgument,
			},
		},
		{
			name:        "Should fail on broken recorder",
			req:         grpcapi.NewGetGaugeReq("Alloc"),
			recorderErr: entity.ErrUnexpected,
			expected: result{
				code: codes.Internal,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)

			m := new(services.RecorderMock)
			m.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
				Return(tc.recorderRV, tc.recorderErr)

			client, closer := createTestServer(t, m)
			t.Cleanup(closer)

			resp, err := client.Get(context.Background(), tc.req)

			status, ok := status.FromError(err)
			require.True(ok)
			require.Equal(tc.expected.code, status.Code())

			if tc.expected.code == codes.OK {
				requireEqual(t, tc.expected.body, resp)
			}
		})
	}
}

func TestBatchUpdate(t *testing.T) {
	batchReq := []*grpcapi.MetricReq{
		grpcapi.NewUpdateCounterReq("PollCount", 10),
		grpcapi.NewUpdateGaugeReq("Alloc", 11.23),
	}

	tt := []struct {
		name        string
		data        []*grpcapi.MetricReq
		recorderErr error
		expected    codes.Code
	}{
		{
			name:     "Should handle list of different metrics",
			data:     batchReq,
			expected: codes.OK,
		},
		{
			name:     "Should fail on empty list",
			data:     make([]*grpcapi.MetricReq, 0),
			expected: codes.InvalidArgument,
		},
		{
			name: "Should fail in unknown metric kind found in list",
			data: []*grpcapi.MetricReq{
				{Id: "xxx", Mtype: "unknown"},
			},
			expected: codes.Unimplemented,
		},
		{
			name:        "Should fail if recorder is broken",
			data:        batchReq,
			recorderErr: entity.ErrUnexpected,
			expected:    codes.Internal,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			m := new(services.RecorderMock)
			m.On("PushList", mock.Anything, mock.Anything).Return(tc.recorderErr)

			client, closer := createTestServer(t, m)
			t.Cleanup(closer)

			req := &grpcapi.BatchUpdateRequest{Data: tc.data}
			_, err := client.BatchUpdate(context.Background(), req)

			status, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tc.expected, status.Code())
		})
	}
}
