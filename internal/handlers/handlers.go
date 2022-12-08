package handlers

import (
	"html/template"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/alkurbatov/metrics-collector/internal/logging"
	"github.com/alkurbatov/metrics-collector/internal/services"
)

type metricsResource struct {
	view     *template.Template
	recorder services.Recorder
}

func newMetricsResource(viewsPath string, recorder services.Recorder) metricsResource {
	view := loadViewTemplate(viewsPath + "/metrics.html")

	return metricsResource{view: view, recorder: recorder}
}

func (h metricsResource) Update(w http.ResponseWriter, r *http.Request) {
	kind := chi.URLParam(r, "kind")
	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")

	if err := validateMetricName(name); err != nil {
		resp := buildResponse(http.StatusBadRequest, err.Error())
		logging.Log.Error(resp)
		http.Error(w, resp, http.StatusBadRequest)
		return
	}

	code := http.StatusOK
	var err error

	switch kind {
	case "counter":
		if err = h.recorder.PushCounter(name, value); err != nil {
			code = http.StatusBadRequest
		}

	case "gauge":
		if err = h.recorder.PushGauge(name, value); err != nil {
			code = http.StatusBadRequest
		}

	default:
		err = errMetricNotImplemented
		code = http.StatusNotImplemented
	}

	if err != nil {
		resp := buildResponse(code, err.Error())
		logging.Log.Error(resp)
		http.Error(w, resp, code)
		return
	}

	logging.Log.Info(codeToResponse(http.StatusOK))
}

func (h metricsResource) Get(w http.ResponseWriter, r *http.Request) {
	kind := chi.URLParam(r, "kind")
	name := chi.URLParam(r, "name")

	if err := validateMetricName(name); err != nil {
		resp := buildResponse(http.StatusBadRequest, err.Error())
		logging.Log.Error(resp)
		http.Error(w, resp, http.StatusBadRequest)
		return
	}

	code := http.StatusOK
	var err error

	switch kind {
	case "counter", "gauge":
		record, ok := h.recorder.GetRecord(kind, name)
		if !ok {
			err = errMetricNotFound
			code = http.StatusNotFound
			break
		}

		if _, err := io.WriteString(w, record.Value.String()); err != nil {
			code = http.StatusInternalServerError
		}

	default:
		err = errMetricNotImplemented
		code = http.StatusNotImplemented
	}

	if err != nil {
		resp := buildResponse(code, err.Error())
		logging.Log.Error(resp)
		http.Error(w, resp, code)
		return
	}

	logging.Log.Info(codeToResponse(http.StatusOK))
}

func (h metricsResource) List(w http.ResponseWriter, r *http.Request) {
	records := h.recorder.ListRecords()
	if err := h.view.Execute(w, records); err != nil {
		code := http.StatusInternalServerError
		resp := buildResponse(code, err.Error())
		logging.Log.Error(resp)
		http.Error(w, resp, code)
		return
	}

	logging.Log.Info(codeToResponse(http.StatusOK))
}
