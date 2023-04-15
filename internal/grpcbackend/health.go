package grpcbackend

import (
	"context"
	"errors"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/services"
	"github.com/alkurbatov/metrics-collector/pkg/grpcapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// HealthServer verifies current health status of the service.
type HealthServer struct {
	grpcapi.UnimplementedHealthServer
	healthcheck services.HealthCheck
}

// NewHealthServer creates new instance of gRPC serving Health API and attaches it to the server.
func NewHealthServer(server *grpc.Server, healthcheck services.HealthCheck) {
	h := &HealthServer{healthcheck: healthcheck}

	grpcapi.RegisterHealthServer(server, h)
}

// Ping verifies connection to the database.
func (h HealthServer) Ping(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	err := h.healthcheck.CheckStorage(ctx)
	if err == nil {
		return new(emptypb.Empty), nil
	}

	if errors.Is(err, entity.ErrHealthCheckNotSupported) {
		return nil, status.Errorf(codes.Unimplemented, err.Error())
	}

	return nil, status.Errorf(codes.Internal, err.Error())
}
