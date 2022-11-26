package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/alkurbatov/metrics-collector/internal/logging"
	"github.com/alkurbatov/metrics-collector/internal/services"
)

type UpdateMetricHandler struct {
	Recorder services.Recorder
}

func (h UpdateMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		if err = h.Recorder.PushCounter(name, value); err != nil {
			code = http.StatusBadRequest
		}

	case "gauge":
		if err = h.Recorder.PushGauge(name, value); err != nil {
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

type GetMetricHandler struct {
	Recorder services.Recorder
}

func (h GetMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		record, ok := h.Recorder.GetRecord(kind, name)
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

type RootHandler struct {
	Recorder services.Recorder
}

func (h RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := `<html>
    <head>
        <title>Collected metrics</title>
    </head>
    <style rel="stylesheet" type="text/css">
        table, td, th {
            border: 1px solid black;
            border-collapse: collapse;
        }

        table {
            width: 100%;
        }

        th {
            text-align: center;
            background-color: lightgray;
        }

        td {
            text-align: left;
        }
    </style>
    <body>
        <table>
            <tr>
                <th>Name</th>
                <th>Kind</th>
                <th>Value</th>
            </tr>
        <tbody>
`

	for _, record := range h.Recorder.ListRecords() {
		resp += "<tr>\n"
		resp += fmt.Sprintf("<td>%s</td>\n", record.Name)
		resp += fmt.Sprintf("<td>%s</td>\n", record.Value.Kind())
		resp += fmt.Sprintf("<td>%s</td>\n", record.Value.String())
		resp += "<tr>\n"
	}

	resp += `
            </tbody>
        </table>
    </body>
</html>
`

	if _, err := io.WriteString(w, resp); err != nil {
		code := http.StatusInternalServerError
		resp := buildResponse(code, err.Error())
		logging.Log.Error(resp)
		http.Error(w, resp, code)
		return
	}

	logging.Log.Info(codeToResponse(http.StatusOK))
}
