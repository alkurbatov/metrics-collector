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
		return lis.Dial() //nolint: wrapcheck
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
			name: "Should push counter",
			req: &grpcapi.MetricReq{
				Id:    "PollCount",
				Mtype: metrics.KindCounter,
				Delta: 10,
			},
			recorderRV: storage.Record{Name: "PollCount", Value: metrics.Counter(10)},
			expected:   codes.OK,
		},
		{
			name: "Should push gauge",
			req: &grpcapi.MetricReq{
				Id:    "Alloc",
				Mtype: metrics.KindGauge,
				Value: 13.123,
			},
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
			name: "Should fail on counter with invalid name",
			req: &grpcapi.MetricReq{
				Id:    "X)",
				Mtype: metrics.KindCounter,
				Delta: 10,
			},
			expected: codes.InvalidArgument,
		},
		{
			name: "Should fail on gauge with invalid name",
			req: &grpcapi.MetricReq{
				Id:    "X;",
				Mtype: metrics.KindGauge,
				Value: 13.123,
			},
			expected: codes.InvalidArgument,
		},
		{
			name: "Should fail on broken recorder",
			req: &grpcapi.MetricReq{
				Id:    "fail",
				Mtype: metrics.KindGauge,
				Value: 13,
			},
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
			req:        &grpcapi.GetMetricRequest{Id: "PollCount", Mtype: metrics.KindCounter},
			recorderRV: storage.Record{Name: "PollCount", Value: metrics.Counter(10)},
			expected: result{
				code: codes.OK,
				body: &grpcapi.MetricReq{
					Id:    "PollCount",
					Mtype: metrics.KindCounter,
					Delta: 10,
				},
			},
		},
		{
			name:       "Should get gauge",
			req:        &grpcapi.GetMetricRequest{Id: "Alloc", Mtype: metrics.KindGauge},
			recorderRV: storage.Record{Name: "Alloc", Value: metrics.Gauge(11.345)},
			expected: result{
				code: codes.OK,
				body: &grpcapi.MetricReq{
					Id:    "Alloc",
					Mtype: metrics.KindGauge,
					Value: 11.345,
				},
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
			req:         &grpcapi.GetMetricRequest{Id: "unknown", Mtype: metrics.KindCounter},
			recorderErr: entity.ErrMetricNotFound,
			expected: result{
				code: codes.NotFound,
			},
		},
		{
			name:        "Should fail on unknown gauge",
			req:         &grpcapi.GetMetricRequest{Id: "unknown", Mtype: metrics.KindGauge},
			recorderErr: entity.ErrMetricNotFound,
			expected: result{
				code: codes.NotFound,
			},
		},
		{
			name: "Should fail on counter with invalid name",
			req:  &grpcapi.GetMetricRequest{Id: "X)", Mtype: metrics.KindCounter},
			expected: result{
				code: codes.InvalidArgument,
			},
		},
		{
			name: "Should fail on gauge with invalid name",
			req:  &grpcapi.GetMetricRequest{Id: "X;", Mtype: metrics.KindGauge},
			expected: result{
				code: codes.InvalidArgument,
			},
		},
		{
			name:        "Should fail on broken recorder",
			req:         &grpcapi.GetMetricRequest{Id: "Alloc", Mtype: metrics.KindGauge},
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
		{Id: "PollCount", Mtype: metrics.KindCounter, Delta: 10},
		{Id: "Alloc", Mtype: metrics.KindGauge, Value: 11.23},
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
