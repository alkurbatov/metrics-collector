package handlers

import (
	"net/http"

	"github.com/alkurbatov/metrics-collector/internal/middleware"
	"github.com/alkurbatov/metrics-collector/internal/services"
	"github.com/go-chi/chi/v5"
)

func Router(viewsPath string, recorder services.Recorder) http.Handler {
	metrics := newMetricsResource(viewsPath, recorder)

	r := chi.NewRouter()

	r.Use(middleware.RequestsLogger)

	r.Get("/", metrics.List)
	r.Get("/value/{kind}/{name}", metrics.Get)
	r.Post("/update/{kind}/{name}/{value}", metrics.Update)

	return r
}
