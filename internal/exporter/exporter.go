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
	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/alkurbatov/metrics-collector/internal/schema"
	"github.com/alkurbatov/metrics-collector/internal/security"
)

type BatchExporter struct {
	// Fully qualified HTTP URL of metrics collector.
	baseURL string

	client *http.Client

	// Entity to sign requests.
	// If set to nil, requests will not be signed.
	signer *security.Signer

	// Internal buffer to store requests.
	buffer []schema.MetricReq

	// Error happened during one of previous method calls.
	// If at least one error occurred, further calls are noop.
	err error
}

func NewBatchExporter(collectorAddress string, secret security.Secret) *BatchExporter {
	baseURL := "http://" + collectorAddress
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	var signer *security.Signer
	if len(secret) > 0 {
		signer = security.NewSigner(secret)
	}

	return &BatchExporter{
		baseURL: baseURL,
		client:  client,
		signer:  signer,
		buffer:  make([]schema.MetricReq, 0),
		err:     nil,
	}
}

func (h *BatchExporter) Add(req schema.MetricReq) *BatchExporter {
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
	payload, err := json.Marshal(h.buffer)
	if err != nil {
		return err
	}

	var compressedReq bytes.Buffer

	compressor, err := gzip.NewWriterLevel(&compressedReq, gzip.BestCompression)
	if err != nil {
		return err
	}

	if _, err = compressor.Write(payload); err != nil {
		return err
	}

	if err = compressor.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, h.baseURL+"/updates", &compressedReq)
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

	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return entity.HTTPError(resp.StatusCode, respBody)
	}

	return nil
}

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

func SendMetrics(ctx context.Context, collectorAddress string, secret security.Secret, stats *metrics.Metrics) error {
	// NB (alkurbatov): Take snapshot to avoid possible races.
	snapshot := *stats

	batch := NewBatchExporter(collectorAddress, secret)

	batch.
		Add(schema.NewUpdateGaugeReq("CPUutilization1", snapshot.Process.CPUutilization1)).
		Add(schema.NewUpdateGaugeReq("TotalMemory", snapshot.Process.TotalMemory)).
		Add(schema.NewUpdateGaugeReq("FreeMemory", snapshot.Process.FreeMemory))

	batch.
		Add(schema.NewUpdateGaugeReq("Alloc", snapshot.Runtime.Alloc)).
		Add(schema.NewUpdateGaugeReq("BuckHashSys", snapshot.Runtime.BuckHashSys)).
		Add(schema.NewUpdateGaugeReq("Frees", snapshot.Runtime.Frees)).
		Add(schema.NewUpdateGaugeReq("GCCPUFraction", snapshot.Runtime.GCCPUFraction)).
		Add(schema.NewUpdateGaugeReq("GCSys", snapshot.Runtime.GCSys)).
		Add(schema.NewUpdateGaugeReq("HeapAlloc", snapshot.Runtime.HeapAlloc)).
		Add(schema.NewUpdateGaugeReq("HeapIdle", snapshot.Runtime.HeapIdle)).
		Add(schema.NewUpdateGaugeReq("HeapInuse", snapshot.Runtime.HeapInuse)).
		Add(schema.NewUpdateGaugeReq("HeapObjects", snapshot.Runtime.HeapObjects)).
		Add(schema.NewUpdateGaugeReq("HeapReleased", snapshot.Runtime.HeapReleased)).
		Add(schema.NewUpdateGaugeReq("HeapSys", snapshot.Runtime.HeapSys)).
		Add(schema.NewUpdateGaugeReq("LastGC", snapshot.Runtime.LastGC)).
		Add(schema.NewUpdateGaugeReq("Lookups", snapshot.Runtime.Lookups)).
		Add(schema.NewUpdateGaugeReq("MCacheInuse", snapshot.Runtime.MCacheInuse)).
		Add(schema.NewUpdateGaugeReq("MCacheSys", snapshot.Runtime.MCacheSys)).
		Add(schema.NewUpdateGaugeReq("MSpanInuse", snapshot.Runtime.MSpanInuse)).
		Add(schema.NewUpdateGaugeReq("MSpanSys", snapshot.Runtime.MSpanSys)).
		Add(schema.NewUpdateGaugeReq("Mallocs", snapshot.Runtime.Mallocs)).
		Add(schema.NewUpdateGaugeReq("NextGC", snapshot.Runtime.NextGC)).
		Add(schema.NewUpdateGaugeReq("NumForcedGC", snapshot.Runtime.NumForcedGC)).
		Add(schema.NewUpdateGaugeReq("NumGC", snapshot.Runtime.NumGC)).
		Add(schema.NewUpdateGaugeReq("OtherSys", snapshot.Runtime.OtherSys)).
		Add(schema.NewUpdateGaugeReq("PauseTotalNs", snapshot.Runtime.PauseTotalNs)).
		Add(schema.NewUpdateGaugeReq("StackInuse", snapshot.Runtime.StackInuse)).
		Add(schema.NewUpdateGaugeReq("StackSys", snapshot.Runtime.StackSys)).
		Add(schema.NewUpdateGaugeReq("Sys", snapshot.Runtime.Sys)).
		Add(schema.NewUpdateGaugeReq("TotalAlloc", snapshot.Runtime.TotalAlloc))

	batch.
		Add(schema.NewUpdateGaugeReq("RandomValue", snapshot.RandomValue))

	batch.
		Add(schema.NewUpdateCounterReq("PollCount", snapshot.PollCount))

	if err := batch.Send(ctx).Error(); err != nil {
		return err
	}

	stats.PollCount -= snapshot.PollCount

	return nil
}
