// Package handlers implements REST API for metrics collector server.
package handlers

// @Title Metrics collector API
// @Description Service for storing metrics data.
// @Version 1.0

// @Contact.name  Alexander Kurbatov
// @Contact.email sir.alkurbatov@yandex.ru

// @Tag.name Metrics
// @Tag.description "Metrics API"

// @Tag.name Healthcheck
// @Tag.description "API to inspect service health state"

import (
	"html/template"
	"net/http"

	// Import pregenerated OpenAPI (Swagger) documentation.
	_ "github.com/alkurbatov/metrics-collector/docs/api"
	"github.com/alkurbatov/metrics-collector/internal/compression"
	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/logging"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/alkurbatov/metrics-collector/internal/services"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

func Router(
	address entity.NetAddress,
	view *template.Template,
	recorder services.Recorder,
	healthcheck services.HealthCheck,
	signer *security.Signer,
) http.Handler {
	metrics := newMetricsResource(view, recorder, signer)
	probe := newLivenessProbe(healthcheck)

	r := chi.NewRouter()

	r.Use(logging.RequestsLogger)
	r.Use(middleware.StripSlashes)
	r.Use(compression.DecompressRequest)
	r.Use(compression.CompressResponse)

	r.Get("/", metrics.List)

	r.Post("/value", metrics.GetJSON)
	r.Get("/value/{kind}/{name}", metrics.Get)

	r.Post("/update", metrics.UpdateJSON)
	r.Post("/updates", metrics.BatchUpdateJSON)
	r.Post("/update/{kind}/{name}/{value}", metrics.Update)

	r.Get("/ping", probe.Ping)

	r.Get("/docs/*", httpSwagger.Handler(
		httpSwagger.URL("http://"+address.String()+"/docs/doc.json"),
	))

	return r
}
