// Package exporter provides means to export collected metrics
// using one of supported transports.
package exporter

import (
	"context"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/alkurbatov/metrics-collector/pkg/metrics"
	"github.com/rs/zerolog/log"
)

// Exporter sends collected metrics to metrics collector in single batch request.
type Exporter interface {
	Add(name string, value metrics.Metric) Exporter
	Send(ctx context.Context) Exporter
	Error() error
	Reset()
	Close() error
}

// New create new instance of Exporter working over specified transport.
func New(
	transport string,
	collectorAddress entity.NetAddress,
	secret security.Secret,
	publicKey security.PublicKey,
) Exporter {
	switch transport {
	case entity.TransportHTTP:
		return NewHTTPExporter(collectorAddress, secret, publicKey)

	case entity.TransportGRPC:
		return NewGRPCExporter(collectorAddress, secret)

	default:
		log.Warn().Msgf("Unknown transport type %s provided, fallback to %s", transport, entity.TransportHTTP)
		return NewHTTPExporter(collectorAddress, secret, publicKey)
	}
}
