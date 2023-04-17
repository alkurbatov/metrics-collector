package httpbackend

import (
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/alkurbatov/metrics-collector/internal/services"
	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/alkurbatov/metrics-collector/internal/validators"
	"github.com/alkurbatov/metrics-collector/pkg/metrics"
)

type metricsResource struct {
	view     *template.Template
	recorder services.Recorder
	signer   *security.Signer
}

func parseUpdateMetricReqList(r *http.Request, signer *security.Signer) ([]storage.Record, error) {
	req := make([]metrics.MetricReq, 0)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}

	rv := make([]storage.Record, len(req))

	for i := range req {
		record, err := toRecord(r.Context(), &req[i], signer)
		if err != nil {
			return nil, err
		}

		rv[i] = record
	}

	if len(rv) == 0 {
		return nil, entity.ErrIncompleteRequest
	}

	return rv, nil
}

func newMetricsResource(
	view *template.Template,
	recorder services.Recorder,
	signer *security.Signer,
) metricsResource {
	return metricsResource{view: view, recorder: recorder, signer: signer}
}

// Update godoc
// @Tags Metrics
// @Router /update/{type}/{name}/{value} [post]
// @Summary Push metric data.
// @ID metrics_update
// @Produce plain
// @Param type path string true "Metrics type (e.g. `counter`, `gauge`)."
// @Param name path string true "Metrics name."
// @Param value path string true "Metrics value, must be convertable to `int64` or `float64`."
// @Success 200 {string} string
// @Failure 400 {string} string http.StatusBadRequest
// @Failure 500 {string} string http.StatusInternalServerError
// @Failure 501 {string} string "Metric type is not supported"
func (h metricsResource) Update(w http.ResponseWriter, r *http.Request) {
	req := metrics.MetricReq{
		ID:    chi.URLParam(r, "name"),
		MType: chi.URLParam(r, "kind"),
	}
	rawValue := chi.URLParam(r, "value")
	ctx := r.Context()

	switch req.MType {
	case metrics.KindCounter:
		delta, err := metrics.ToCounter(rawValue)
		if err != nil {
			writeErrorResponse(ctx, w, http.StatusBadRequest, err)
			return
		}

		req.Delta = &delta

	case metrics.KindGauge:
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

// UpdateJSON godoc
// @Tags Metrics
// @Router /update [post]
// @Summary Push metric data as JSON
// @ID metrics_json_update
// @Accept  json
// @Param request body metrics.MetricReq true "Request parameters."
// @Success 200 {object} metrics.MetricReq
// @Failure 400 {string} string http.StatusBadRequest
// @Failure 500 {string} string http.StatusInternalServerError
// @Failure 501 {string} string "Metric type is not supported"
func (h metricsResource) UpdateJSON(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := new(metrics.MetricReq)
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
		hash, err := h.signer.CalculateRecordSignature(recorded)
		if err != nil {
			writeErrorResponse(ctx, w, http.StatusInternalServerError, err)
			return
		}

		resp.Hash = hash
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		writeErrorResponse(ctx, w, http.StatusInternalServerError, err)
		return
	}
}

// BatchUpdateJSON godoc
// @Tags Metrics
// @Router /updates [post]
// @Summary Push list of metrics data as JSON
// @ID metrics_json_update_list
// @Accept  json
// @Param request body []metrics.MetricReq true "List of metrics to update."
// @Success 200
// @Failure 400 {string} string http.StatusBadRequest
// @Failure 500 {string} string http.StatusInternalServerError
// @Failure 501 {string} string "Metric type is not supported"
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

// Get godoc
// @Tags Metrics
// @Router /value/{type}/{name} [get]
// @Summary Get metrics value as string
// @ID metrics_info
// @Produce plain
// @Param type path string true "Metrics type (e.g. `counter`, `gauge`)."
// @Param name path string true "Metrics name."
// @Success 200 {string} string
// @Failure 400 {string} string http.StatusBadRequest
// @Failure 404 {string} string "Metric not found"
// @Failure 500 {string} string http.StatusInternalServerError
// @Failure 501 {string} string "Metric type is not supported"
func (h metricsResource) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	kind := chi.URLParam(r, "kind")
	name := chi.URLParam(r, "name")

	if err := validators.ValidateMetricName(name, kind); err != nil {
		writeErrorResponse(ctx, w, http.StatusBadRequest, err)
		return
	}

	if err := validators.ValidateMetricKind(kind); err != nil {
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

// GetJSON godoc
// @Tags Metrics
// @Router /value [post]
// @Summary Get metrics value as JSON
// @ID metrics_json_info
// @Accept  json
// @Produce json
// @Param request body metrics.MetricReq true "Request parameters: `id` and `type` are required."
// @Success 200 {object} metrics.MetricReq
// @Failure 400 {string} string http.StatusBadRequest
// @Failure 404 {string} string "Metric not found"
// @Failure 500 {string} string http.StatusInternalServerError
// @Failure 501 {string} string "Metric type is not supported"
func (h metricsResource) GetJSON(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := new(metrics.MetricReq)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		writeErrorResponse(ctx, w, http.StatusBadRequest, err)
		return
	}

	if err := validators.ValidateMetricName(req.ID, req.MType); err != nil {
		writeErrorResponse(ctx, w, http.StatusBadRequest, err)
		return
	}

	if err := validators.ValidateMetricKind(req.MType); err != nil {
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
		hash, err := h.signer.CalculateRecordSignature(record)
		if err != nil {
			writeErrorResponse(ctx, w, http.StatusInternalServerError, err)
			return
		}

		resp.Hash = hash
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		writeErrorResponse(ctx, w, http.StatusInternalServerError, err)
		return
	}
}

// List godoc
// @Tags Metrics
// @Router / [get]
// @Summary Get HTML page with full list of stored metrics
// @ID metrics_list
// @Produce html
// @Success 200
// @Failure 500 {string} string http.StatusInternalServerError
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

// Ping godoc
// @Tags Healthcheck
// @Router /ping [get]
// @Summary Verify connection to the database
// @ID health_info
// @Success 200
// @Failure 500 {string} string "Connection is broken"
// @Failure 501 {string} string "Server is not configured to use database"
func (h livenessProbe) Ping(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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
