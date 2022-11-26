package main

import (
	"net/http"

	"github.com/alkurbatov/metrics-collector/internal/app"
	"github.com/alkurbatov/metrics-collector/internal/handlers"
	"github.com/alkurbatov/metrics-collector/internal/logging"
	"github.com/alkurbatov/metrics-collector/internal/services"
)

func main() {
	app := app.NewServer()
	recorder := services.NewMetricsRecorder(app)

	logging.Log.Info("Listening on " + app.Config.ListenAddress)

	if err := http.ListenAndServe(app.Config.ListenAddress, handlers.Router(recorder)); err != nil {
		logging.Log.Fatal(err)
	}
}
