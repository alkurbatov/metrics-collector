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
	"github.com/alkurbatov/metrics-collector/internal/logging"
	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/alkurbatov/metrics-collector/internal/schema"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/alkurbatov/metrics-collector/internal/services"
)

type metricsResource struct {
	view     *template.Template
	recorder services.Recorder
	signer   *security.Signer
}

func parse(r *http.Request, signer *security.Signer) (*schema.MetricReq, error) {
	req := new(schema.MetricReq)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return nil, err
	}

	if err := schema.ValidateMetricName(req.ID, req.MType); err != nil {
		return nil, err
	}

	if signer == nil {
		return req, nil
	}

	valid, err := signer.VerifySignature(req)
	if err != nil {
		// NB (alkurbatov): We don't want to give any hints to potential attacker,
		// but still want to debug implementation errors. Thus, the error is only logged.
		logging.Log.Error(err)
	}

	if err != nil || !valid {
		return nil, entity.ErrInvalidSignature
	}

	return req, nil
}

func writeErrorResponse(w http.ResponseWriter, code int, err error) {
	resp := buildResponse(code, err.Error())
	logging.Log.Error(resp)
	http.Error(w, resp, code)
}

func newMetricsResource(viewsPath string, recorder services.Recorder, signer *security.Signer) metricsResource {
	view := loadViewTemplate(viewsPath + "/metrics.html")

	return metricsResource{view: view, recorder: recorder, signer: signer}
}

func (h metricsResource) Update(w http.ResponseWriter, r *http.Request) {
	kind := chi.URLParam(r, "kind")
	name := chi.URLParam(r, "name")
	rawValue := chi.URLParam(r, "value")

	if err := schema.ValidateMetricName(name, kind); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	switch kind {
	case entity.Counter:
		value, err := metrics.ToCounter(rawValue)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, err)
			return
		}

		newDelta, err := h.recorder.PushCounter(r.Context(), name, value)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		if _, err = io.WriteString(w, newDelta.String()); err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

	case entity.Gauge:
		value, err := metrics.ToGauge(rawValue)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, err)
			return
		}

		newValue, err := h.recorder.PushGauge(r.Context(), name, value)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		if _, err = io.WriteString(w, newValue.String()); err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

	default:
		writeErrorResponse(w, http.StatusNotImplemented, entity.ErrMetricNotImplemented)
		return
	}

	logging.Log.Info(OK())
}

func (h metricsResource) UpdateJSON(w http.ResponseWriter, r *http.Request) {
	data, err := parse(r, h.signer)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	switch data.MType {
	case entity.Counter:
		newDelta, err := h.recorder.PushCounter(r.Context(), data.ID, *data.Delta)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		data.Delta = &newDelta

	case entity.Gauge:
		newValue, err := h.recorder.PushGauge(r.Context(), data.ID, *data.Value)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		data.Value = &newValue

	default:
		writeErrorResponse(w, http.StatusNotImplemented, entity.ErrMetricNotImplemented)
		return
	}

	if h.signer != nil {
		if err := h.signer.SignRequest(data); err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, err)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(data); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	logging.Log.Info(OK())
}

func (h metricsResource) Get(w http.ResponseWriter, r *http.Request) {
	kind := chi.URLParam(r, "kind")
	name := chi.URLParam(r, "name")

	if err := schema.ValidateMetricName(name, kind); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	switch kind {
	case entity.Counter, entity.Gauge:
		record, err := h.recorder.GetRecord(r.Context(), kind, name)
		if err != nil {
			if errors.Is(err, entity.ErrMetricNotFound) {
				writeErrorResponse(w, http.StatusNotFound, err)
				return
			}

			writeErrorResponse(w, http.StatusInternalServerError, err)

			return
		}

		if _, err := io.WriteString(w, record.Value.String()); err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

	default:
		writeErrorResponse(w, http.StatusNotImplemented, entity.ErrMetricNotImplemented)
		return
	}

	logging.Log.Info(OK())
}

func (h metricsResource) GetJSON(w http.ResponseWriter, r *http.Request) {
	data, err := parse(r, nil)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	switch data.MType {
	case entity.Counter:
		record, err := h.recorder.GetRecord(r.Context(), data.MType, data.ID)
		if err != nil {
			if errors.Is(err, entity.ErrMetricNotFound) {
				writeErrorResponse(w, http.StatusNotFound, err)
				return
			}

			writeErrorResponse(w, http.StatusInternalServerError, err)

			return
		}

		delta, ok := record.Value.(metrics.Counter)
		if !ok {
			writeErrorResponse(w, http.StatusNotFound, entity.ErrRecordKindDontMatch)
			return
		}

		data.Delta = &delta

	case entity.Gauge:
		record, err := h.recorder.GetRecord(r.Context(), data.MType, data.ID)
		if err != nil {
			if errors.Is(err, entity.ErrMetricNotFound) {
				writeErrorResponse(w, http.StatusNotFound, err)
				return
			}

			writeErrorResponse(w, http.StatusInternalServerError, err)

			return
		}

		value, ok := record.Value.(metrics.Gauge)
		if !ok {
			writeErrorResponse(w, http.StatusNotFound, entity.ErrRecordKindDontMatch)
			return
		}

		data.Value = &value

	default:
		writeErrorResponse(w, http.StatusNotImplemented, entity.ErrMetricNotImplemented)
		return
	}

	if h.signer != nil {
		if err := h.signer.SignRequest(data); err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, err)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(data); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	logging.Log.Info(OK())
}

func (h metricsResource) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	records, err := h.recorder.ListRecords(r.Context())
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	if err := h.view.Execute(w, records); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	logging.Log.Info(OK())
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
		logging.Log.Info(OK())
		return
	}

	if errors.Is(err, entity.ErrHealthCheckNotSupported) {
		writeErrorResponse(w, http.StatusNotImplemented, err)
		return
	}

	writeErrorResponse(w, http.StatusInternalServerError, err)
}
