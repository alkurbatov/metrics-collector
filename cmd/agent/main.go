package main

import (
	"context"

	"github.com/alkurbatov/metrics-collector/internal/app"
	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/rs/zerolog/log"
)

func main() {
	app := app.NewAgent()
	ctx := context.Background()

	stats := &metrics.Metrics{}
	go app.Poll(ctx, stats)
	go app.Report(ctx, stats)

	log.Info().Msg("Agent has started")
	select {}
}
