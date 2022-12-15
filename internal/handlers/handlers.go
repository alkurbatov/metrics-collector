package handlers

import (
	"encoding/json"
	"html/template"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/alkurbatov/metrics-collector/internal/logging"
	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/alkurbatov/metrics-collector/internal/schema"
	"github.com/alkurbatov/metrics-collector/internal/services"
)

type metricsResource struct {
	view     *template.Template
	recorder services.Recorder
}

func writeErrorResponse(w http.ResponseWriter, code int, err error) {
	resp := buildResponse(code, err.Error())
	logging.Log.Error(resp)
	http.Error(w, resp, code)
}

func newMetricsResource(viewsPath string, recorder services.Recorder) metricsResource {
	view := loadViewTemplate(viewsPath + "/metrics.html")

	return metricsResource{view: view, recorder: recorder}
}

func (h metricsResource) Update(w http.ResponseWriter, r *http.Request) {
	kind := chi.URLParam(r, "kind")
	name := chi.URLParam(r, "name")
	rawValue := chi.URLParam(r, "value")

	if err := validateMetricName(name); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	switch kind {
	case "counter":
		value, err := metrics.ToCounter(rawValue)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, err)
			return
		}

		newDelta, err := h.recorder.PushCounter(name, value)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		if _, err = io.WriteString(w, newDelta.String()); err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

	case "gauge":
		value, err := metrics.ToGauge(rawValue)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, err)
			return
		}

		newValue, err := h.recorder.PushGauge(name, value)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		if _, err = io.WriteString(w, newValue.String()); err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

	default:
		writeErrorResponse(w, http.StatusNotImplemented, errMetricNotImplemented)
		return
	}

	logging.Log.Info(codeToResponse(http.StatusOK))
}

func (h metricsResource) UpdateJSON(w http.ResponseWriter, r *http.Request) {
	data := &schema.MetricReq{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	if err := validateMetricName(data.ID); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	switch data.MType {
	case "counter":
		newDelta, err := h.recorder.PushCounter(data.ID, *data.Delta)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		data.Delta = &newDelta

	case "gauge":
		newValue, err := h.recorder.PushGauge(data.ID, *data.Value)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		data.Value = &newValue

	default:
		writeErrorResponse(w, http.StatusNotImplemented, errMetricNotImplemented)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(data); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	logging.Log.Info(codeToResponse(http.StatusOK))
}

func (h metricsResource) Get(w http.ResponseWriter, r *http.Request) {
	kind := chi.URLParam(r, "kind")
	name := chi.URLParam(r, "name")

	if err := validateMetricName(name); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	switch kind {
	case "counter", "gauge":
		record, ok := h.recorder.GetRecord(kind, name)
		if !ok {
			writeErrorResponse(w, http.StatusNotFound, errMetricNotFound)
			return
		}

		if _, err := io.WriteString(w, record.Value.String()); err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

	default:
		writeErrorResponse(w, http.StatusNotImplemented, errMetricNotImplemented)
		return
	}

	logging.Log.Info(codeToResponse(http.StatusOK))
}

func (h metricsResource) GetJSON(w http.ResponseWriter, r *http.Request) {
	data := &schema.MetricReq{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	if err := validateMetricName(data.ID); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	switch data.MType {
	case "counter":
		record, ok := h.recorder.GetRecord(data.MType, data.ID)
		if !ok {
			writeErrorResponse(w, http.StatusNotFound, errMetricNotFound)
			return
		}

		delta, ok := record.Value.(metrics.Counter)
		if !ok {
			writeErrorResponse(w, http.StatusNotFound, errRecordKindDontMatch)
			return
		}

		data.Delta = &delta

	case "gauge":
		record, ok := h.recorder.GetRecord(data.MType, data.ID)
		if !ok {
			writeErrorResponse(w, http.StatusNotFound, errMetricNotFound)
			return
		}

		value, ok := record.Value.(metrics.Gauge)
		if !ok {
			writeErrorResponse(w, http.StatusNotFound, errRecordKindDontMatch)
			return
		}

		data.Value = &value

	default:
		writeErrorResponse(w, http.StatusNotImplemented, errMetricNotImplemented)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(data); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	logging.Log.Info(codeToResponse(http.StatusOK))
}

func (h metricsResource) List(w http.ResponseWriter, r *http.Request) {
	records := h.recorder.ListRecords()
	if err := h.view.Execute(w, records); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	logging.Log.Info(codeToResponse(http.StatusOK))
}
