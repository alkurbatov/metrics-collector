package main

import (
	"context"

	"github.com/alkurbatov/metrics-collector/internal/app"
	"github.com/alkurbatov/metrics-collector/internal/logging"
	"github.com/alkurbatov/metrics-collector/internal/metrics"
)

func main() {
	app := app.NewAgent()
	ctx := context.Background()

	stats := &metrics.Metrics{}
	go app.Poll(ctx, stats)
	go app.Report(ctx, stats)

	logging.Log.Info("Agent has started")
	select {}
}
