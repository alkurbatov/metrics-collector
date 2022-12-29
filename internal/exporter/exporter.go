package exporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/alkurbatov/metrics-collector/internal/logging"
	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/alkurbatov/metrics-collector/internal/schema"
	"github.com/alkurbatov/metrics-collector/internal/security"
)

type HTTPExporter struct {
	baseURL string
	client  *http.Client
	signer  *security.Signer
	err     error
}

func NewExporter(collectorAddress string, secret security.Secret) HTTPExporter {
	baseURL := fmt.Sprintf("http://%s", collectorAddress)
	client := &http.Client{Timeout: 2 * time.Second}

	var signer *security.Signer
	if len(secret) > 0 {
		signer = security.NewSigner(secret)
	}

	return HTTPExporter{
		baseURL: baseURL,
		client:  client,
		signer:  signer,
		err:     nil,
	}
}

func (h *HTTPExporter) doExport(req *schema.MetricReq) *HTTPExporter {
	logging.Log.WithField("type", req.MType).Info("Update " + req.ID)

	if h.signer != nil {
		if err := h.signer.SignRequest(req); err != nil {
			h.err = err
			return h
		}
	}

	payload, err := json.Marshal(req)
	if err != nil {
		h.err = err
		return h
	}

	resp, err := h.client.Post(h.baseURL+"/update", "Content-Type: application/json", bytes.NewReader(payload))
	if err != nil {
		h.err = err
		return h
	}

	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)

	if err != nil {
		h.err = err
		return h
	}

	if resp.StatusCode != http.StatusOK {
		h.err = fmt.Errorf("metrics export failed: (%d)", resp.StatusCode)
		return h
	}

	return h
}

func (h *HTTPExporter) exportGauge(name string, value metrics.Gauge) *HTTPExporter {
	if h.err != nil {
		return h
	}

	req := schema.NewUpdateGaugeReq(name, value)

	return h.doExport(&req)
}

func (h *HTTPExporter) exportCounter(name string, value metrics.Counter) *HTTPExporter {
	if h.err != nil {
		return h
	}

	req := schema.NewUpdateCounterReq(name, value)

	return h.doExport(&req)
}

func SendMetrics(collectorAddress string, secret security.Secret, stats metrics.Metrics) error {
	exporter := NewExporter(collectorAddress, secret)

	exporter.
		exportGauge("Alloc", stats.Memory.Alloc).
		exportGauge("BuckHashSys", stats.Memory.BuckHashSys).
		exportGauge("Frees", stats.Memory.Frees).
		exportGauge("GCCPUFraction", stats.Memory.GCCPUFraction).
		exportGauge("GCSys", stats.Memory.GCSys).
		exportGauge("HeapAlloc", stats.Memory.HeapAlloc).
		exportGauge("HeapIdle", stats.Memory.HeapIdle).
		exportGauge("HeapInuse", stats.Memory.HeapInuse).
		exportGauge("HeapObjects", stats.Memory.HeapObjects).
		exportGauge("HeapReleased", stats.Memory.HeapReleased).
		exportGauge("HeapSys", stats.Memory.HeapSys).
		exportGauge("LastGC", stats.Memory.LastGC).
		exportGauge("Lookups", stats.Memory.Lookups).
		exportGauge("MCacheInuse", stats.Memory.MCacheInuse).
		exportGauge("MCacheSys", stats.Memory.MCacheSys).
		exportGauge("MSpanInuse", stats.Memory.MSpanInuse).
		exportGauge("MSpanSys", stats.Memory.MSpanSys).
		exportGauge("Mallocs", stats.Memory.Mallocs).
		exportGauge("NextGC", stats.Memory.NextGC).
		exportGauge("NumForcedGC", stats.Memory.NumForcedGC).
		exportGauge("NumGC", stats.Memory.NumGC).
		exportGauge("OtherSys", stats.Memory.OtherSys).
		exportGauge("PauseTotalNs", stats.Memory.PauseTotalNs).
		exportGauge("StackInuse", stats.Memory.StackInuse).
		exportGauge("StackSys", stats.Memory.StackSys).
		exportGauge("Sys", stats.Memory.Sys).
		exportGauge("TotalAlloc", stats.Memory.TotalAlloc)

	exporter.
		exportGauge("RandomValue", stats.RandomValue)

	exporter.
		exportCounter("PollCount", stats.PollCount)

	return exporter.err
}
