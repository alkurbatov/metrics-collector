package exporter

import (
	"context"
	"fmt"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/pkg/grpcapi"
	"github.com/alkurbatov/metrics-collector/pkg/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ServerConnection *grpc.ClientConn

// GRPCExporter sends collected metrics to metrics collector in single batch request.
type GRPCExporter struct {
	// Address and port of server providing gRPC API.
	endpoint entity.NetAddress

	// Internal buffer to store requests.
	buffer []*grpcapi.MetricReq

	// Error happened during one of previous method calls.
	// If at least one error occurred, further calls are noop.
	err error
}

func NewGRPCExporter(endpoint entity.NetAddress) *GRPCExporter {
	return &GRPCExporter{endpoint: endpoint}
}

// Add a metric to internal buffer.
func (g *GRPCExporter) Add(name string, value metrics.Metric) Exporter {
	if g.err != nil {
		return g
	}

	var req *grpcapi.MetricReq
	switch v := value.(type) {
	case metrics.Counter:
		req = grpcapi.NewUpdateCounterReq(name, v)

	case metrics.Gauge:
		req = grpcapi.NewUpdateGaugeReq(name, v)

	default:
		g.err = entity.MetricNotImplementedError(value.Kind())
		return g
	}

	g.buffer = append(g.buffer, req)

	return g
}

func (g *GRPCExporter) Error() error {
	if g.err == nil {
		return nil
	}

	return fmt.Errorf("metrics export failed: %w", g.err)
}

// Send metrics stored in internal buffer to metrics collector in single batch request.
func (g *GRPCExporter) Send(ctx context.Context) Exporter {
	if g.err != nil {
		return g
	}

	if _ServerConnection == nil {
		_ServerConnection, g.err = grpc.DialContext(
			ctx,
			g.endpoint.String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if g.err != nil {
			return g
		}
	}

	client := grpcapi.NewMetricsClient(_ServerConnection)

	if len(g.buffer) == 0 {
		g.err = entity.ErrIncompleteRequest
		return g
	}

	req := &grpcapi.BatchUpdateRequest{Data: g.buffer}
	_, g.err = client.BatchUpdate(ctx, req)

	return g
}
