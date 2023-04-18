package exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/alkurbatov/metrics-collector/internal/compression"
	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/alkurbatov/metrics-collector/pkg/metrics"
)

// HTTPExporter sends collected metrics to metrics collector in single batch request.
type HTTPExporter struct {
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

func NewHTTPExporter(
	collectorAddress entity.NetAddress,
	secret security.Secret,
	publicKey security.PublicKey,
) *HTTPExporter {
	baseURL := "http://" + collectorAddress.String()
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	var signer *security.Signer
	if len(secret) > 0 {
		signer = security.NewSigner(secret)
	}

	return &HTTPExporter{
		baseURL:   baseURL,
		client:    client,
		signer:    signer,
		buffer:    make([]metrics.MetricReq, 0),
		publicKey: publicKey,
		err:       nil,
	}
}

func (h *HTTPExporter) Error() error {
	if h.err == nil {
		return nil
	}

	return fmt.Errorf("metrics export failed: %w", h.err)
}

// Add a metric to internal buffer.
func (h *HTTPExporter) Add(name string, value metrics.Metric) Exporter {
	if h.err != nil {
		return h
	}

	var req metrics.MetricReq
	switch v := value.(type) {
	case metrics.Counter:
		req = metrics.NewUpdateCounterReq(name, v)

	case metrics.Gauge:
		req = metrics.NewUpdateGaugeReq(name, v)

	default:
		h.err = entity.MetricNotImplementedError(value.Kind())
		return h
	}

	if h.signer != nil {
		hash, err := h.signer.CalculateSignature(name, value)
		if err != nil {
			h.err = err
			return h
		}

		req.Hash = hash
	}

	h.buffer = append(h.buffer, req)

	return h
}

func (h *HTTPExporter) doSend(ctx context.Context) error {
	jsonReq, err := json.Marshal(h.buffer)
	if err != nil {
		return err
	}

	payload, err := compression.Pack(jsonReq)
	if err != nil {
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

	clientIP, err := getOutboundIP()
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("X-Real-IP", clientIP.String())

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
func (h *HTTPExporter) Send(ctx context.Context) Exporter {
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

// Reset reset state of exporter to initial.
// This doesn't affected the underlying connection.
func (h *HTTPExporter) Reset() {
	h.buffer = make([]metrics.MetricReq, 0)
	h.err = nil
}

// Close of HTTP exporter is no-op.
// Required by the Exporter interface.
func (h *HTTPExporter) Close() error {
	return nil
}
