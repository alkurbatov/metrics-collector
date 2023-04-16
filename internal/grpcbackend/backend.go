// Package grpcbackend implements gRPC API for metrics collector server.
package grpcbackend

import (
	"net"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/grpcserver"
	"github.com/alkurbatov/metrics-collector/internal/logging"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/alkurbatov/metrics-collector/internal/services"
	"google.golang.org/grpc"
)

func New(
	address entity.NetAddress,
	recorder services.Recorder,
	healthcheck services.HealthCheck,
	trustedSubnet *net.IPNet,
) *grpcserver.Server {
	interceptors := make([]grpc.UnaryServerInterceptor, 0, 2)
	interceptors = append(interceptors, logging.UnaryRequestsInterceptor)

	if trustedSubnet != nil {
		interceptors = append(interceptors, security.UnaryRequestsFilter(trustedSubnet))
	}

	grpcSrv := grpcserver.New(
		address,
		grpc.ChainUnaryInterceptor(interceptors...),
	)
	NewHealthServer(grpcSrv.Instance(), healthcheck)
	NewMetricsServer(grpcSrv.Instance(), recorder)

	return grpcSrv
}
