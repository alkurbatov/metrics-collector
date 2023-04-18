package exporter

import (
	"context"
	"fmt"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/alkurbatov/metrics-collector/pkg/grpcapi"
	"github.com/alkurbatov/metrics-collector/pkg/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// GRPCExporter sends collected metrics to metrics collector in single batch request.
type GRPCExporter struct {
	// Address and port of server providing gRPC API.
	endpoint entity.NetAddress

	conn *grpc.ClientConn

	// Entity to sign requests.
	// If set to nil, requests will not be signed.
	signer *security.Signer

	// Internal buffer to store requests.
	buffer []*grpcapi.MetricReq

	// Error happened during one of previous method calls.
	// If at least one error occurred, further calls are noop.
	err error
}

func NewGRPCExporter(
	endpoint entity.NetAddress,
	secret security.Secret,
) *GRPCExporter {
	var signer *security.Signer
	if len(secret) > 0 {
		signer = security.NewSigner(secret)
	}

	return &GRPCExporter{
		endpoint: endpoint,
		signer:   signer,
	}
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

	if g.signer != nil {
		hash, err := g.signer.CalculateSignature(name, value)
		if err != nil {
			g.err = err
			return g
		}

		req.Hash = hash
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

	if g.conn == nil {
		g.conn, g.err = grpc.DialContext(
			ctx,
			g.endpoint.String(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if g.err != nil {
			return g
		}
	}

	if len(g.buffer) == 0 {
		g.err = entity.ErrIncompleteRequest
		return g
	}

	clientIP, err := getOutboundIP()
	if err != nil {
		g.err = err
		return g
	}

	md := metadata.New(map[string]string{"x-real-ip": clientIP.String()})
	ctx = metadata.NewOutgoingContext(ctx, md)
	client := grpcapi.NewMetricsClient(g.conn)

	req := &grpcapi.BatchUpdateRequest{Data: g.buffer}
	_, g.err = client.BatchUpdate(ctx, req)

	return g
}

// Reset reset state of exporter to initial.
// This doesn't affected the underlying connection.
func (g *GRPCExporter) Reset() {
	g.buffer = make([]*grpcapi.MetricReq, 0)
	g.err = nil
}

// Close gracefully finishes gRPC client connection.
func (g *GRPCExporter) Close() error {
	if g.conn == nil {
		return nil
	}

	return g.conn.Close()
}
