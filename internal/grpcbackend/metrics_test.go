package grpcbackend_test

import (
	"context"
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/alkurbatov/metrics-collector/internal/services"
	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/alkurbatov/metrics-collector/pkg/grpcapi"
	"github.com/alkurbatov/metrics-collector/pkg/metrics"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
)

func TestUpdateMetric(t *testing.T) {
	tt := []struct {
		name        string
		req         *grpcapi.MetricReq
		clientKey   security.Secret
		serverKey   security.Secret
		recorderRV  storage.Record
		recorderErr error
		expected    codes.Code
	}{
		{
			name:       "Push unsigned counter",
			req:        grpcapi.NewUpdateCounterReq("PollCount", 10),
			recorderRV: storage.Record{Name: "PollCount", Value: metrics.Counter(10)},
			expected:   codes.OK,
		},
		{
			name:       "Push unsigned gauge",
			req:        grpcapi.NewUpdateGaugeReq("Alloc", 13.123),
			recorderRV: storage.Record{Name: "Alloc", Value: metrics.Gauge(13.123)},
			expected:   codes.OK,
		},
		{
			name:       "Push signed counter",
			req:        grpcapi.NewUpdateCounterReq("PollCount", 10),
			clientKey:  "abc",
			serverKey:  "abc",
			recorderRV: storage.Record{Name: "PollCount", Value: metrics.Counter(10)},
			expected:   codes.OK,
		},
		{
			name:       "Push signed gauge",
			req:        grpcapi.NewUpdateGaugeReq("Alloc", 13.123),
			clientKey:  "abc",
			serverKey:  "abc",
			recorderRV: storage.Record{Name: "Alloc", Value: metrics.Gauge(13.123)},
			expected:   codes.OK,
		},
		{
			name:       "Push counter fails if signature doesn't match",
			req:        grpcapi.NewUpdateCounterReq("PollCount", 10),
			clientKey:  "abc",
			serverKey:  "xxx",
			recorderRV: storage.Record{Name: "PollCount", Value: metrics.Counter(10)},
			expected:   codes.InvalidArgument,
		},
		{
			name:       "Push gauge fails if signature doesn't match",
			req:        grpcapi.NewUpdateGaugeReq("Alloc", 13.123),
			clientKey:  "abc",
			serverKey:  "xxx",
			recorderRV: storage.Record{Name: "Alloc", Value: metrics.Gauge(13.123)},
			expected:   codes.InvalidArgument,
		},
		{
			name: "Push metric fails if type is not known",
			req: &grpcapi.MetricReq{
				Id:    "X",
				Mtype: "unknown",
			},
			expected: codes.Unimplemented,
		},
		{
			name:     "Push counter fails if name is invalid",
			req:      grpcapi.NewUpdateCounterReq("X)", 10),
			expected: codes.InvalidArgument,
		},
		{
			name:     "Push gauge fails if name is invalid",
			req:      grpcapi.NewUpdateGaugeReq("X;", 13.123),
			expected: codes.InvalidArgument,
		},
		{
			name:        "Push metrics fails if recorder is broken",
			req:         grpcapi.NewUpdateGaugeReq("fail", 13),
			recorderErr: entity.ErrUnexpected,
			expected:    codes.Internal,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			m := new(services.RecorderMock)
			m.On("Push", mock.Anything, mock.AnythingOfType("Record")).Return(tc.recorderRV, tc.recorderErr)

			conn, closer := createTestServer(t, m, nil, tc.serverKey)
			t.Cleanup(closer)

			if len(tc.clientKey) != 0 {
				signer := security.NewSigner(tc.clientKey)

				hash, err := signer.CalculateRecordSignature(tc.recorderRV)
				require.NoError(t, err)

				tc.req.Hash = hash
			}

			client := grpcapi.NewMetricsClient(conn)
			resp, err := client.Update(context.Background(), tc.req)

			requireEqualCode(t, tc.expected, err)

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
		serverKey   security.Secret
		recorderRV  storage.Record
		recorderErr error
		expected    result
	}{
		{
			name:       "Get unsigned counter",
			req:        grpcapi.NewGetCounterReq("PollCount"),
			recorderRV: storage.Record{Name: "PollCount", Value: metrics.Counter(10)},
			expected: result{
				code: codes.OK,
				body: grpcapi.NewUpdateCounterReq("PollCount", 10),
			},
		},
		{
			name:       "Get unsigned gauge",
			req:        grpcapi.NewGetGaugeReq("Alloc"),
			recorderRV: storage.Record{Name: "Alloc", Value: metrics.Gauge(11.345)},
			expected: result{
				code: codes.OK,
				body: grpcapi.NewUpdateGaugeReq("Alloc", 11.345),
			},
		},
		{
			name:       "Get signed counter",
			req:        grpcapi.NewGetCounterReq("PollCount"),
			serverKey:  "abc",
			recorderRV: storage.Record{Name: "PollCount", Value: metrics.Counter(10)},
			expected: result{
				code: codes.OK,
				body: &grpcapi.MetricReq{
					Id:    "PollCount",
					Mtype: metrics.KindCounter,
					Delta: 10,
					Hash:  "0833001195f2e062140968e0c00dd44f00eb9a0b309aedc464817f904b244c8a",
				},
			},
		},
		{
			name:       "Get signed gauge",
			req:        grpcapi.NewGetGaugeReq("Alloc"),
			serverKey:  "abc",
			recorderRV: storage.Record{Name: "Alloc", Value: metrics.Gauge(11.345)},
			expected: result{
				code: codes.OK,
				body: &grpcapi.MetricReq{
					Id:    "Alloc",
					Mtype: metrics.KindGauge,
					Value: 11.345,
					Hash:  "2d32037265fd3547d65d4f51d69d8ea53490bef6e924fa2cfe2e4045ad50527d",
				},
			},
		},

		{
			name: "Get metric fails if type is unknown",
			req:  &grpcapi.GetMetricRequest{Id: "Alloc", Mtype: "unknown"},
			expected: result{
				code: codes.Unimplemented,
			},
		},
		{
			name:        "Get counter fails if counter has unknown ID",
			req:         grpcapi.NewGetCounterReq("unknown"),
			recorderErr: entity.ErrMetricNotFound,
			expected: result{
				code: codes.NotFound,
			},
		},
		{
			name:        "Get metric fails if gauge has unknown ID",
			req:         grpcapi.NewGetGaugeReq("unknown"),
			recorderErr: entity.ErrMetricNotFound,
			expected: result{
				code: codes.NotFound,
			},
		},
		{
			name: "Get counter fails if name is invalid",
			req:  grpcapi.NewGetCounterReq("X)"),
			expected: result{
				code: codes.InvalidArgument,
			},
		},
		{
			name: "Get gauge fails if name is invalid",
			req:  grpcapi.NewGetGaugeReq("X;"),
			expected: result{
				code: codes.InvalidArgument,
			},
		},
		{
			name:        "Get metric fails if recorder is broken",
			req:         grpcapi.NewGetGaugeReq("Alloc"),
			recorderErr: entity.ErrUnexpected,
			expected: result{
				code: codes.Internal,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			m := new(services.RecorderMock)
			m.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
				Return(tc.recorderRV, tc.recorderErr)

			conn, closer := createTestServer(t, m, nil, tc.serverKey)
			t.Cleanup(closer)

			client := grpcapi.NewMetricsClient(conn)
			resp, err := client.Get(context.Background(), tc.req)

			requireEqualCode(t, tc.expected.code, err)

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

	batchResp := []storage.Record{
		{Name: "Alloc", Value: metrics.Gauge(11.23)},
		{Name: "PollCount", Value: metrics.Counter(10)},
	}

	type expected struct {
		code     codes.Code
		response []*grpcapi.MetricReq
	}

	tt := []struct {
		name        string
		data        []*grpcapi.MetricReq
		clientKey   security.Secret
		serverKey   security.Secret
		recorderRv  []storage.Record
		recorderErr error
		expected    expected
	}{
		{
			name:       "Batch update handles signed list of different metrics",
			data:       batchReq,
			clientKey:  "abc",
			serverKey:  "abc",
			recorderRv: batchResp,
			expected: expected{
				code: codes.OK,
				response: []*grpcapi.MetricReq{
					grpcapi.NewUpdateGaugeReq("Alloc", 11.23),
					grpcapi.NewUpdateCounterReq("PollCount", 10),
				},
			},
		},
		{
			name:       "Batch update handles unsigned list of different metrics",
			data:       batchReq,
			recorderRv: batchResp,
			expected: expected{
				code: codes.OK,
				response: []*grpcapi.MetricReq{
					grpcapi.NewUpdateGaugeReq("Alloc", 11.23),
					grpcapi.NewUpdateCounterReq("PollCount", 10),
				},
			},
		},
		{
			name: "Batch update fails on empty list",
			data: make([]*grpcapi.MetricReq, 0),
			expected: expected{
				code: codes.InvalidArgument,
			},
		},
		{
			name: "Batch update fails if unknown metric kind found in list",
			data: []*grpcapi.MetricReq{
				{Id: "xxx", Mtype: "unknown"},
			},
			expected: expected{
				code: codes.Unimplemented,
			},
		},
		{
			name:        "Batch update fails if recorder is broken",
			data:        batchReq,
			recorderErr: entity.ErrUnexpected,
			expected: expected{
				code: codes.Internal,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			m := new(services.RecorderMock)
			m.On("PushList", mock.Anything, mock.Anything).Return(tc.recorderRv, tc.recorderErr)

			conn, closer := createTestServer(t, m, nil, "")
			t.Cleanup(closer)

			if len(tc.clientKey) > 0 {
				signer := security.NewSigner(tc.clientKey)

				hash, err := signer.CalculateSignature(tc.data[0].Id, metrics.Counter(tc.data[0].Delta))
				require.NoError(t, err)
				tc.data[0].Hash = hash

				hash, err = signer.CalculateSignature(tc.data[1].Id, metrics.Gauge(tc.data[1].Value))
				require.NoError(t, err)
				tc.data[1].Hash = hash
			}

			client := grpcapi.NewMetricsClient(conn)
			req := &grpcapi.BatchUpdateRequest{Data: tc.data}
			resp, err := client.BatchUpdate(context.Background(), req)

			requireEqualCode(t, tc.expected.code, err)
			if tc.expected.code == codes.OK {
				require.Equal(t, tc.expected.response, resp.Data)
			}
		})
	}
}
