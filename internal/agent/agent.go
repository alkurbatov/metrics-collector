// Package agent implements metrics gathering agent.
package agent

import (
	"context"
	"errors"
	"time"

	"github.com/alkurbatov/metrics-collector/internal/exporter"
	"github.com/alkurbatov/metrics-collector/internal/logging"
	"github.com/alkurbatov/metrics-collector/internal/monitoring"
	"github.com/alkurbatov/metrics-collector/internal/recovery"
	"github.com/rs/zerolog/log"
)

type Agent struct {
	config Config
	stats  monitoring.Metrics
}

func New() *Agent {
	cfg := NewConfig()

	logging.Setup(cfg.Debug)
	log.Info().Msg(cfg.String())

	return &Agent{config: cfg}
}

func (app *Agent) poll(ctx context.Context, stats *monitoring.Metrics) {
	ticker := time.NewTicker(app.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			func() {
				defer recovery.TryRecover()

				log.Info().Msg("Gathering application metrics")

				taskCtx, cancel := context.WithTimeout(ctx, app.config.PollTimeout)
				defer cancel()

				err := stats.Poll(taskCtx)

				if errors.Is(taskCtx.Err(), context.DeadlineExceeded) {
					log.Error().Dur("deadline", app.config.PollTimeout).Msg("metrics polling exceeded deadline")
					return
				}

				if err != nil {
					log.Error().Err(err).Msg("")
					return
				}

				log.Info().Msg("Metrics gathered")
			}()

		case <-ctx.Done():
			log.Info().Msg("Shutdown metrics gathering")
			return
		}
	}
}

func (app *Agent) report(ctx context.Context, stats *monitoring.Metrics) {
	ticker := time.NewTicker(app.config.ReportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			func() {
				defer recovery.TryRecover()

				log.Info().Msg("Sending application metrics")

				taskCtx, cancel := context.WithTimeout(ctx, app.config.ExportTimeout)
				defer cancel()

				err := exporter.SendMetrics(taskCtx, app.config.CollectorAddress, app.config.Secret, stats)

				if errors.Is(err, context.DeadlineExceeded) {
					log.Error().Dur("deadline", app.config.PollTimeout).Msg("metrics exporting exceeded deadline")
					return
				}

				if err != nil {
					log.Error().Err(err).Msg("")
					return
				}

				log.Info().Msg("Metrics successfully sent")
			}()

		case <-ctx.Done():
			log.Info().Msg("Shutdown metrics sending")
			return
		}
	}
}

func (app *Agent) Serve(ctx context.Context) {
	go app.poll(ctx, &app.stats)
	go app.report(ctx, &app.stats)
}
