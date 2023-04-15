package grpcbackend

import (
	"context"
	"errors"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/services"
	"github.com/alkurbatov/metrics-collector/internal/validators"
	"github.com/alkurbatov/metrics-collector/pkg/grpcapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MetricsServer allows to store and retrieve metrics.
type MetricsServer struct {
	grpcapi.UnimplementedMetricsServer

	recorder services.Recorder
}

// NewMetricsServer creates new instance of gRPC serving Metrics API and attaches it to the server.
func NewMetricsServer(server *grpc.Server, recorder services.Recorder) {
	s := &MetricsServer{recorder: recorder}

	grpcapi.RegisterMetricsServer(server, s)
}

// Update pushes metric data to the server.
func (s MetricsServer) Update(ctx context.Context, req *grpcapi.MetricReq) (*grpcapi.MetricReq, error) {
	record, err := toRecord(req)
	if err != nil {
		if errors.Is(err, entity.ErrMetricNotImplemented) {
			return nil, status.Errorf(codes.Unimplemented, err.Error())
		}

		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	recorded, err := s.recorder.Push(ctx, record)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return toMetricReq(recorded), nil
}

// Get metrics value.
func (s MetricsServer) Get(ctx context.Context, req *grpcapi.GetMetricRequest) (*grpcapi.MetricReq, error) {
	if err := validators.ValidateMetricName(req.Id, req.Mtype); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := validators.ValidateMetricKind(req.Mtype); err != nil {
		return nil, status.Errorf(codes.Unimplemented, err.Error())
	}

	record, err := s.recorder.Get(ctx, req.Mtype, req.Id)
	if err != nil {
		if errors.Is(err, entity.ErrMetricNotFound) {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}

		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return toMetricReq(record), nil
}
