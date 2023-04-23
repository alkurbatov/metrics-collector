package grpcbackend

import (
	"context"
	"errors"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/security"
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
	signer   *security.Signer
}

// NewMetricsServer creates new instance of gRPC serving Metrics API and attaches it to the server.
func NewMetricsServer(server *grpc.Server, recorder services.Recorder, signer *security.Signer) {
	s := &MetricsServer{recorder: recorder, signer: signer}

	grpcapi.RegisterMetricsServer(server, s)
}

// Update pushes metric data to the server.
func (s MetricsServer) Update(ctx context.Context, req *grpcapi.MetricReq) (*grpcapi.MetricReq, error) {
	record, err := toRecord(ctx, req, s.signer)
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

	resp, err := toMetricReq(recorded, s.signer)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return resp, nil
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

	resp, err := toMetricReq(record, s.signer)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return resp, nil
}

// BatchUpdate pushes list of metrics data.
func (s MetricsServer) BatchUpdate(
	ctx context.Context,
	req *grpcapi.BatchUpdateRequest,
) (*grpcapi.BatchUpdateResponse, error) {
	records, err := toRecordsList(ctx, req, s.signer)
	if err != nil {
		if errors.Is(err, entity.ErrMetricNotImplemented) {
			return nil, status.Errorf(codes.Unimplemented, err.Error())
		}

		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	records, err = s.recorder.PushList(ctx, records)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	data, err := toMetricReqList(records, s.signer)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &grpcapi.BatchUpdateResponse{Data: data}, nil
}
