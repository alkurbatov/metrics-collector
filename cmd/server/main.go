package main

import (
	"net/http"

	"github.com/alkurbatov/metrics-collector/internal/app"
	"github.com/alkurbatov/metrics-collector/internal/handlers"
	"github.com/alkurbatov/metrics-collector/internal/logging"
	"github.com/alkurbatov/metrics-collector/internal/middleware"
)

type Middleware func(http.Handler) http.Handler

func attachMiddleware(h http.Handler) http.Handler {
	common := [...]Middleware{middleware.RequestsLogger}

	for _, middleware := range common {
		h = middleware(h)
	}
	return h
}

func router(app *app.Server) http.Handler {
	r := http.NewServeMux()

	r.Handle(
		"/",
		attachMiddleware(http.NotFoundHandler()),
	)
	r.Handle(
		"/update/",
		attachMiddleware(http.StripPrefix("/update/", handlers.UpdateMetricHandler{App: app})),
	)

	return r
}

func main() {
	app := app.NewServer()
	logging.Log.Info("Listening on " + app.Config.ListenAddress)

	if err := http.ListenAndServe(app.Config.ListenAddress, router(app)); err != nil {
		logging.Log.Fatal(err)
	}
}
