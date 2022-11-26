package handlers

import (
	"net/http"

	"github.com/alkurbatov/metrics-collector/internal/middleware"
	"github.com/alkurbatov/metrics-collector/internal/services"
	"github.com/go-chi/chi/v5"
)

func Router(recorder services.Recorder) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestsLogger)

	r.Method("GET", "/", RootHandler{Recorder: recorder})
	r.Method("GET", "/value/{kind}/{name}", GetMetricHandler{Recorder: recorder})
	r.Method("POST", "/update/{kind}/{name}/{value}", UpdateMetricHandler{Recorder: recorder})

	return r
}
