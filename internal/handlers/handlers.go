package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/alkurbatov/metrics-collector/internal/schema"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/alkurbatov/metrics-collector/internal/services"
	"github.com/alkurbatov/metrics-collector/internal/storage"
)

type metricsResource struct {
	view     *template.Template
	recorder services.Recorder
	signer   *security.Signer
}

func parseUpdateMetricReqList(r *http.Request, signer *security.Signer) ([]storage.Record, error) {
	req := make([]schema.MetricReq, 0)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err //nolint: wrapcheck
	}

	rv := make([]storage.Record, 0)

	for i := range req {
		record, err := toRecord(r.Context(), &req[i], signer)
		if err != nil {
			return nil, err
		}

		rv = append(rv, record)
	}

	if len(rv) == 0 {
		return nil, entity.ErrIncompleteRequest
	}

	return rv, nil
}

func newMetricsResource(viewsPath string, recorder services.Recorder, signer *security.Signer) metricsResource {
	view := loadViewTemplate(viewsPath + "/metrics.html")

	return metricsResource{view: view, recorder: recorder, signer: signer}
}

func (h metricsResource) Update(w http.ResponseWriter, r *http.Request) {
	req := schema.MetricReq{
		ID:    chi.URLParam(r, "name"),
		MType: chi.URLParam(r, "kind"),
	}
	rawValue := chi.URLParam(r, "value")
	ctx := r.Context()

	switch req.MType {
	case entity.Counter:
		delta, err := metrics.ToCounter(rawValue)
		if err != nil {
			writeErrorResponse(ctx, w, http.StatusBadRequest, err)
			return
		}

		req.Delta = &delta

	case entity.Gauge:
		value, err := metrics.ToGauge(rawValue)
		if err != nil {
			writeErrorResponse(ctx, w, http.StatusBadRequest, err)
			return
		}

		req.Value = &value
	}

	record, err := toRecord(ctx, &req, nil)
	if err != nil {
		if errors.Is(err, entity.ErrMetricNotImplemented) {
			writeErrorResponse(ctx, w, http.StatusNotImplemented, err)
			return
		}

		writeErrorResponse(ctx, w, http.StatusBadRequest, err)

		return
	}

	recorded, err := h.recorder.Push(r.Context(), record)
	if err != nil {
		writeErrorResponse(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if _, err = io.WriteString(w, recorded.Value.String()); err != nil {
		writeErrorResponse(r.Context(), w, http.StatusInternalServerError, err)
		return
	}
}

func (h metricsResource) UpdateJSON(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := new(schema.MetricReq)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		writeErrorResponse(ctx, w, http.StatusBadRequest, err)
		return
	}

	record, err := toRecord(ctx, req, h.signer)
	if err != nil {
		if errors.Is(err, entity.ErrMetricNotImplemented) {
			writeErrorResponse(ctx, w, http.StatusNotImplemented, err)
			return
		}

		writeErrorResponse(ctx, w, http.StatusBadRequest, err)

		return
	}

	recorded, err := h.recorder.Push(r.Context(), record)
	if err != nil {
		writeErrorResponse(ctx, w, http.StatusInternalServerError, err)
		return
	}

	resp := toMetricReq(recorded)

	if h.signer != nil {
		if err := h.signer.SignRequest(&resp); err != nil {
			writeErrorResponse(ctx, w, http.StatusInternalServerError, err)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		writeErrorResponse(ctx, w, http.StatusInternalServerError, err)
		return
	}
}

func (h metricsResource) BatchUpdateJSON(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := parseUpdateMetricReqList(r, h.signer)
	if err != nil {
		if errors.Is(err, entity.ErrMetricNotImplemented) {
			writeErrorResponse(ctx, w, http.StatusNotImplemented, err)
			return
		}

		writeErrorResponse(ctx, w, http.StatusBadRequest, err)

		return
	}

	if err := h.recorder.PushList(r.Context(), req); err != nil {
		writeErrorResponse(ctx, w, http.StatusInternalServerError, err)
		return
	}
}

func (h metricsResource) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	kind := chi.URLParam(r, "kind")
	name := chi.URLParam(r, "name")

	if err := schema.ValidateMetricName(name, kind); err != nil {
		writeErrorResponse(ctx, w, http.StatusBadRequest, err)
		return
	}

	if err := schema.ValidateMetricKind(kind); err != nil {
		writeErrorResponse(ctx, w, http.StatusNotImplemented, err)
		return
	}

	record, err := h.recorder.Get(r.Context(), kind, name)
	if err != nil {
		if errors.Is(err, entity.ErrMetricNotFound) {
			writeErrorResponse(ctx, w, http.StatusNotFound, err)
			return
		}

		writeErrorResponse(ctx, w, http.StatusInternalServerError, err)

		return
	}

	if _, err := io.WriteString(w, record.Value.String()); err != nil {
		writeErrorResponse(ctx, w, http.StatusInternalServerError, err)
		return
	}
}

func (h metricsResource) GetJSON(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := new(schema.MetricReq)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		writeErrorResponse(ctx, w, http.StatusBadRequest, err)
		return
	}

	if err := schema.ValidateMetricName(req.ID, req.MType); err != nil {
		writeErrorResponse(ctx, w, http.StatusBadRequest, err)
		return
	}

	if err := schema.ValidateMetricKind(req.MType); err != nil {
		writeErrorResponse(ctx, w, http.StatusNotImplemented, err)
		return
	}

	record, err := h.recorder.Get(r.Context(), req.MType, req.ID)
	if err != nil {
		if errors.Is(err, entity.ErrMetricNotFound) {
			writeErrorResponse(ctx, w, http.StatusNotFound, err)
			return
		}

		writeErrorResponse(ctx, w, http.StatusInternalServerError, err)

		return
	}

	resp := toMetricReq(record)

	if h.signer != nil {
		if err := h.signer.SignRequest(&resp); err != nil {
			writeErrorResponse(ctx, w, http.StatusInternalServerError, err)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		writeErrorResponse(ctx, w, http.StatusInternalServerError, err)
		return
	}
}

func (h metricsResource) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	records, err := h.recorder.List(r.Context())
	if err != nil {
		writeErrorResponse(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err := h.view.Execute(w, records); err != nil {
		writeErrorResponse(ctx, w, http.StatusInternalServerError, err)
		return
	}
}

type livenessProbe struct {
	healthcheck services.HealthCheck
}

func newLivenessProbe(healthcheck services.HealthCheck) livenessProbe {
	return livenessProbe{healthcheck: healthcheck}
}

func (h livenessProbe) Ping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err := h.healthcheck.CheckStorage(ctx)
	if err == nil {
		return
	}

	if errors.Is(err, entity.ErrHealthCheckNotSupported) {
		writeErrorResponse(ctx, w, http.StatusNotImplemented, err)
		return
	}

	writeErrorResponse(ctx, w, http.StatusInternalServerError, err)
}
