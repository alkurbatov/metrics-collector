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
	router := handlers.Router("./web/views", recorder)

	logging.Log.Info(app.Config)

	if err := http.ListenAndServe(app.Config.ListenAddress, router); err != nil {
		logging.Log.Fatal(err)
	}
}
