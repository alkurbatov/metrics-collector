// Package exporter provides means to export collected metrics
// using one of supported transports.
package exporter

import (
	"context"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/monitoring"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/alkurbatov/metrics-collector/pkg/metrics"
)

// Exporter sends collected metrics to metrics collector in single batch request.
type Exporter interface {
	Add(name string, value metrics.Metric) Exporter
	Send(ctx context.Context) Exporter
	Error() error
}

// NewExporter create new instance of Exporter working over specified transport.
func NewExporter(
	transport string,
	collectorAddress entity.NetAddress,
	secret security.Secret,
	publicKey security.PublicKey,
) (Exporter, error) {
	switch transport {
	case entity.TransportHTTP:
		return NewHTTPExporter(collectorAddress, secret, publicKey), nil

	case entity.TransportGRPC:
		return NewGRPCExporter(collectorAddress), nil

	default:
		return nil, entity.TransportNotSupportedError(transport)
	}
}

// SendMetrics exports collected metrics in single batch request.
func SendMetrics(
	ctx context.Context,
	transport string,
	collectorAddress entity.NetAddress,
	secret security.Secret,
	publicKey security.PublicKey,
	stats *monitoring.Metrics,
) error {
	// NB (alkurbatov): Take snapshot to avoid possible races.
	snapshot := *stats

	exp, err := NewExporter(transport, collectorAddress, secret, publicKey)
	if err != nil {
		return err
	}

	exp.
		Add("CPUutilization1", snapshot.System.CPUutilization1).
		Add("TotalMemory", snapshot.System.TotalMemory).
		Add("FreeMemory", snapshot.System.FreeMemory)

	exp.
		Add("Alloc", snapshot.Runtime.Alloc).
		Add("BuckHashSys", snapshot.Runtime.BuckHashSys).
		Add("Frees", snapshot.Runtime.Frees).
		Add("GCCPUFraction", snapshot.Runtime.GCCPUFraction).
		Add("GCSys", snapshot.Runtime.GCSys).
		Add("HeapAlloc", snapshot.Runtime.HeapAlloc).
		Add("HeapIdle", snapshot.Runtime.HeapIdle).
		Add("HeapInuse", snapshot.Runtime.HeapInuse).
		Add("HeapObjects", snapshot.Runtime.HeapObjects).
		Add("HeapReleased", snapshot.Runtime.HeapReleased).
		Add("HeapSys", snapshot.Runtime.HeapSys).
		Add("LastGC", snapshot.Runtime.LastGC).
		Add("Lookups", snapshot.Runtime.Lookups).
		Add("MCacheInuse", snapshot.Runtime.MCacheInuse).
		Add("MCacheSys", snapshot.Runtime.MCacheSys).
		Add("MSpanInuse", snapshot.Runtime.MSpanInuse).
		Add("MSpanSys", snapshot.Runtime.MSpanSys).
		Add("Mallocs", snapshot.Runtime.Mallocs).
		Add("NextGC", snapshot.Runtime.NextGC).
		Add("NumForcedGC", snapshot.Runtime.NumForcedGC).
		Add("NumGC", snapshot.Runtime.NumGC).
		Add("OtherSys", snapshot.Runtime.OtherSys).
		Add("PauseTotalNs", snapshot.Runtime.PauseTotalNs).
		Add("StackInuse", snapshot.Runtime.StackInuse).
		Add("StackSys", snapshot.Runtime.StackSys).
		Add("Sys", snapshot.Runtime.Sys).
		Add("TotalAlloc", snapshot.Runtime.TotalAlloc)

	exp.
		Add("RandomValue", snapshot.RandomValue)

	exp.
		Add("PollCount", snapshot.PollCount)

	if err := exp.Send(ctx).Error(); err != nil {
		return err
	}

	stats.PollCount -= snapshot.PollCount

	return nil
}
