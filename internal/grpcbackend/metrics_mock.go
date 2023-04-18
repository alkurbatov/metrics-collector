package grpcbackend

import (
	"context"

	"github.com/alkurbatov/metrics-collector/pkg/grpcapi"
	"github.com/stretchr/testify/mock"
)

type MetricsServerMock struct {
	mock.Mock
	grpcapi.UnimplementedMetricsServer
}

func (m *MetricsServerMock) Update(ctx context.Context, req *grpcapi.MetricReq) (*grpcapi.MetricReq, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*grpcapi.MetricReq), args.Error(1)
}

func (m *MetricsServerMock) Get(ctx context.Context, req *grpcapi.GetMetricRequest) (*grpcapi.MetricReq, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*grpcapi.MetricReq), args.Error(1)
}

func (m *MetricsServerMock) BatchUpdate(
	ctx context.Context,
	req *grpcapi.BatchUpdateRequest,
) (*grpcapi.BatchUpdateResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*grpcapi.BatchUpdateResponse), args.Error(1)
}
