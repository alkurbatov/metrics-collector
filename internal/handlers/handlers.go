package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/alkurbatov/metrics-collector/internal/app"
	"github.com/alkurbatov/metrics-collector/internal/logging"
)

type UpdateMetricHandler struct {
	BaseURL string
	App     *app.Server
}

func (h UpdateMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		resp := codeToResponse(http.StatusMethodNotAllowed)
		logging.Log.Error(resp)
		http.Error(w, resp, http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.String(), "/")

	if len(parts) != 3 {
		resp := codeToResponse(http.StatusNotFound)
		logging.Log.Error(resp)
		http.Error(w, resp, http.StatusNotFound)
		return
	}

	switch parts[0] {
	case "counter":
		req, err := parseUpdateCounterReq(parts...)
		if err != nil {
			resp := buildResponse(http.StatusBadRequest, err.Error())
			logging.Log.Error(resp)
			http.Error(w, resp, http.StatusBadRequest)
			return
		}

		h.App.Storage.PushCounter(req.name, req.value)
		resp := buildResponse(http.StatusOK, fmt.Sprintf("Set %s += %d", req.name, req.value))
		logging.Log.Info(resp)

	case "gauge":
		req, err := parseUpdateGaugeReq(parts...)
		if err != nil {
			resp := buildResponse(http.StatusBadRequest, err.Error())
			logging.Log.Error(resp)
			http.Error(w, resp, http.StatusBadRequest)
			return
		}

		h.App.Storage.PushGauge(req.name, req.value)
		resp := buildResponse(http.StatusOK, fmt.Sprintf("Set %s = %f", req.name, req.value))
		logging.Log.Info(resp)

	default:
		resp := codeToResponse(http.StatusNotImplemented)
		logging.Log.Error(resp)
		http.Error(w, resp, http.StatusNotImplemented)
	}
}
