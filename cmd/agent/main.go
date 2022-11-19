package main

import (
	"context"

	"github.com/alkurbatov/metrics-collector/internal/app"
	"github.com/alkurbatov/metrics-collector/internal/metrics"

	log "github.com/sirupsen/logrus"
)

func main() {
	app := app.NewAgent()
	ctx := context.Background()

	stats := metrics.NewMetrics()
	go app.Poll(ctx, stats)
	go app.Report(ctx, stats)

	log.Info("Agent has started")
	select {}
}
