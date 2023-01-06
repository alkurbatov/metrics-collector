package exporter

import (
	"bytes"
	"compress/gzip"
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

func (h *BatchExporter) doSend() error {
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

func (h *BatchExporter) Send() *BatchExporter {
	if h.err != nil {
		return h
	}

	if len(h.buffer) == 0 {
		h.err = entity.ErrIncompleteRequest
		return h
	}

	h.err = h.doSend()

	return h
}

func SendMetrics(collectorAddress string, secret security.Secret, stats metrics.Metrics) error {
	batch := NewBatchExporter(collectorAddress, secret)

	batch.
		Add(schema.NewUpdateGaugeReq("Alloc", stats.Memory.Alloc)).
		Add(schema.NewUpdateGaugeReq("BuckHashSys", stats.Memory.BuckHashSys)).
		Add(schema.NewUpdateGaugeReq("Frees", stats.Memory.Frees)).
		Add(schema.NewUpdateGaugeReq("GCCPUFraction", stats.Memory.GCCPUFraction)).
		Add(schema.NewUpdateGaugeReq("GCSys", stats.Memory.GCSys)).
		Add(schema.NewUpdateGaugeReq("HeapAlloc", stats.Memory.HeapAlloc)).
		Add(schema.NewUpdateGaugeReq("HeapIdle", stats.Memory.HeapIdle)).
		Add(schema.NewUpdateGaugeReq("HeapInuse", stats.Memory.HeapInuse)).
		Add(schema.NewUpdateGaugeReq("HeapObjects", stats.Memory.HeapObjects)).
		Add(schema.NewUpdateGaugeReq("HeapReleased", stats.Memory.HeapReleased)).
		Add(schema.NewUpdateGaugeReq("HeapSys", stats.Memory.HeapSys)).
		Add(schema.NewUpdateGaugeReq("LastGC", stats.Memory.LastGC)).
		Add(schema.NewUpdateGaugeReq("Lookups", stats.Memory.Lookups)).
		Add(schema.NewUpdateGaugeReq("MCacheInuse", stats.Memory.MCacheInuse)).
		Add(schema.NewUpdateGaugeReq("MCacheSys", stats.Memory.MCacheSys)).
		Add(schema.NewUpdateGaugeReq("MSpanInuse", stats.Memory.MSpanInuse)).
		Add(schema.NewUpdateGaugeReq("MSpanSys", stats.Memory.MSpanSys)).
		Add(schema.NewUpdateGaugeReq("Mallocs", stats.Memory.Mallocs)).
		Add(schema.NewUpdateGaugeReq("NextGC", stats.Memory.NextGC)).
		Add(schema.NewUpdateGaugeReq("NumForcedGC", stats.Memory.NumForcedGC)).
		Add(schema.NewUpdateGaugeReq("NumGC", stats.Memory.NumGC)).
		Add(schema.NewUpdateGaugeReq("OtherSys", stats.Memory.OtherSys)).
		Add(schema.NewUpdateGaugeReq("PauseTotalNs", stats.Memory.PauseTotalNs)).
		Add(schema.NewUpdateGaugeReq("StackInuse", stats.Memory.StackInuse)).
		Add(schema.NewUpdateGaugeReq("StackSys", stats.Memory.StackSys)).
		Add(schema.NewUpdateGaugeReq("Sys", stats.Memory.Sys)).
		Add(schema.NewUpdateGaugeReq("TotalAlloc", stats.Memory.TotalAlloc))

	batch.
		Add(schema.NewUpdateGaugeReq("RandomValue", stats.RandomValue))

	batch.
		Add(schema.NewUpdateCounterReq("PollCount", stats.PollCount))

	return batch.Send().Error()
}
