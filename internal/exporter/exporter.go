// Package exporter provides means to export collected metrics.
package exporter

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/monitoring"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/alkurbatov/metrics-collector/pkg/metrics"
)

// BatchExporter sends collected metrics to metrics collector in single batch request.
type BatchExporter struct {
	// Fully qualified HTTP URL of metrics collector.
	baseURL string

	client *http.Client

	// Entity to sign requests.
	// If set to nil, requests will not be signed.
	signer *security.Signer

	// Internal buffer to store requests.
	buffer []metrics.MetricReq

	// Public key to encrypt data.
	// If not set, the data is sent unencrypted.
	publicKey security.PublicKey

	// Error happened during one of previous method calls.
	// If at least one error occurred, further calls are noop.
	err error
}

func NewBatchExporter(
	collectorAddress entity.NetAddress,
	secret security.Secret,
	publicKey security.PublicKey,
) *BatchExporter {
	baseURL := "http://" + collectorAddress.String()
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	var signer *security.Signer
	if len(secret) > 0 {
		signer = security.NewSigner(secret)
	}

	return &BatchExporter{
		baseURL:   baseURL,
		client:    client,
		signer:    signer,
		buffer:    make([]metrics.MetricReq, 0),
		publicKey: publicKey,
		err:       nil,
	}
}

// Add a metric to internal buffer.
func (h *BatchExporter) Add(req metrics.MetricReq) *BatchExporter {
	if h.err != nil {
		return h
	}

	if h.signer != nil {
		if err := h.signer.SignRequest(&req); err != nil {
			h.err = err
			return h
		}
	}

	h.buffer = append(h.buffer, req)

	return h
}

func (h *BatchExporter) Error() error {
	if h.err == nil {
		return nil
	}

	return fmt.Errorf("metrics export failed: %w", h.err)
}

func (h *BatchExporter) doSend(ctx context.Context) error {
	jsonReq, err := json.Marshal(h.buffer)
	if err != nil {
		return err
	}

	payload := new(bytes.Buffer)

	compressor, err := gzip.NewWriterLevel(payload, gzip.BestCompression)
	if err != nil {
		return err
	}

	if _, err = compressor.Write(jsonReq); err != nil {
		return err
	}

	if err = compressor.Close(); err != nil {
		return err
	}

	if h.publicKey != nil {
		payload, err = security.Encrypt(io.Reader(payload), h.publicKey)
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest(http.MethodPost, h.baseURL+"/updates", payload)
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return entity.HTTPError(resp.StatusCode, respBody)
	}

	return nil
}

// Send metrics stored in internal buffer to metrics collector in single batch request.
func (h *BatchExporter) Send(ctx context.Context) *BatchExporter {
	if h.err != nil {
		return h
	}

	if len(h.buffer) == 0 {
		h.err = entity.ErrIncompleteRequest
		return h
	}

	h.err = h.doSend(ctx)

	return h
}

// SendMetrics exports collected metrics in single batch request.
func SendMetrics(
	ctx context.Context,
	collectorAddress entity.NetAddress,
	secret security.Secret,
	publicKey security.PublicKey,
	stats *monitoring.Metrics,
) error {
	// NB (alkurbatov): Take snapshot to avoid possible races.
	snapshot := *stats

	batch := NewBatchExporter(collectorAddress, secret, publicKey)

	batch.
		Add(metrics.NewUpdateGaugeReq("CPUutilization1", snapshot.System.CPUutilization1)).
		Add(metrics.NewUpdateGaugeReq("TotalMemory", snapshot.System.TotalMemory)).
		Add(metrics.NewUpdateGaugeReq("FreeMemory", snapshot.System.FreeMemory))

	batch.
		Add(metrics.NewUpdateGaugeReq("Alloc", snapshot.Runtime.Alloc)).
		Add(metrics.NewUpdateGaugeReq("BuckHashSys", snapshot.Runtime.BuckHashSys)).
		Add(metrics.NewUpdateGaugeReq("Frees", snapshot.Runtime.Frees)).
		Add(metrics.NewUpdateGaugeReq("GCCPUFraction", snapshot.Runtime.GCCPUFraction)).
		Add(metrics.NewUpdateGaugeReq("GCSys", snapshot.Runtime.GCSys)).
		Add(metrics.NewUpdateGaugeReq("HeapAlloc", snapshot.Runtime.HeapAlloc)).
		Add(metrics.NewUpdateGaugeReq("HeapIdle", snapshot.Runtime.HeapIdle)).
		Add(metrics.NewUpdateGaugeReq("HeapInuse", snapshot.Runtime.HeapInuse)).
		Add(metrics.NewUpdateGaugeReq("HeapObjects", snapshot.Runtime.HeapObjects)).
		Add(metrics.NewUpdateGaugeReq("HeapReleased", snapshot.Runtime.HeapReleased)).
		Add(metrics.NewUpdateGaugeReq("HeapSys", snapshot.Runtime.HeapSys)).
		Add(metrics.NewUpdateGaugeReq("LastGC", snapshot.Runtime.LastGC)).
		Add(metrics.NewUpdateGaugeReq("Lookups", snapshot.Runtime.Lookups)).
		Add(metrics.NewUpdateGaugeReq("MCacheInuse", snapshot.Runtime.MCacheInuse)).
		Add(metrics.NewUpdateGaugeReq("MCacheSys", snapshot.Runtime.MCacheSys)).
		Add(metrics.NewUpdateGaugeReq("MSpanInuse", snapshot.Runtime.MSpanInuse)).
		Add(metrics.NewUpdateGaugeReq("MSpanSys", snapshot.Runtime.MSpanSys)).
		Add(metrics.NewUpdateGaugeReq("Mallocs", snapshot.Runtime.Mallocs)).
		Add(metrics.NewUpdateGaugeReq("NextGC", snapshot.Runtime.NextGC)).
		Add(metrics.NewUpdateGaugeReq("NumForcedGC", snapshot.Runtime.NumForcedGC)).
		Add(metrics.NewUpdateGaugeReq("NumGC", snapshot.Runtime.NumGC)).
		Add(metrics.NewUpdateGaugeReq("OtherSys", snapshot.Runtime.OtherSys)).
		Add(metrics.NewUpdateGaugeReq("PauseTotalNs", snapshot.Runtime.PauseTotalNs)).
		Add(metrics.NewUpdateGaugeReq("StackInuse", snapshot.Runtime.StackInuse)).
		Add(metrics.NewUpdateGaugeReq("StackSys", snapshot.Runtime.StackSys)).
		Add(metrics.NewUpdateGaugeReq("Sys", snapshot.Runtime.Sys)).
		Add(metrics.NewUpdateGaugeReq("TotalAlloc", snapshot.Runtime.TotalAlloc))

	batch.
		Add(metrics.NewUpdateGaugeReq("RandomValue", snapshot.RandomValue))

	batch.
		Add(metrics.NewUpdateCounterReq("PollCount", snapshot.PollCount))

	if err := batch.Send(ctx).Error(); err != nil {
		return err
	}

	stats.PollCount -= snapshot.PollCount

	return nil
}
