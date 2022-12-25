package handlers

import (
	"net/http"

	"github.com/alkurbatov/metrics-collector/internal/compression"
	"github.com/alkurbatov/metrics-collector/internal/logging"
	"github.com/alkurbatov/metrics-collector/internal/services"
	"github.com/go-chi/chi/v5"
)

func Router(viewsPath string, recorder services.Recorder, signer *services.Signer) http.Handler {
	metrics := newMetricsResource(viewsPath, recorder, signer)

	r := chi.NewRouter()

	r.Use(logging.RequestsLogger)
	r.Use(compression.DecompressRequest)
	r.Use(compression.CompressResponse)

	r.Get("/", metrics.List)

	r.Post("/value", metrics.GetJSON)
	r.Post("/value/", metrics.GetJSON)
	r.Get("/value/{kind}/{name}", metrics.Get)

	r.Post("/update", metrics.UpdateJSON)
	r.Post("/update/", metrics.UpdateJSON)
	r.Post("/update/{kind}/{name}/{value}", metrics.Update)

	return r
}
