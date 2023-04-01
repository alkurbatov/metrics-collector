// Package agent implements metrics gathering agent.
package agent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alkurbatov/metrics-collector/internal/config"
	"github.com/alkurbatov/metrics-collector/internal/exporter"
	"github.com/alkurbatov/metrics-collector/internal/monitoring"
	"github.com/alkurbatov/metrics-collector/internal/recovery"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/rs/zerolog/log"
)

type Agent struct {
	config *config.Agent
	stats  *monitoring.Metrics

	// Public key used to encrypt agent -> server communications.
	// If the key is nil, communications are not encrypted.
	publicKey security.PublicKey
}

func New(cfg *config.Agent) (*Agent, error) {
	var (
		key security.PublicKey
		err error
	)

	if len(cfg.PublicKeyPath) != 0 {
		key, err = security.NewPublicKey(cfg.PublicKeyPath)
		if err != nil {
			return nil, fmt.Errorf("agent - New - security.NewPublicKey: %w", err)
		}
	}

	stats := monitoring.NewMetrics()

	return &Agent{
		config:    cfg,
		stats:     stats,
		publicKey: key,
	}, nil
}

func (app *Agent) poll(ctx context.Context) {
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

				err := app.stats.Poll(taskCtx)

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

func (app *Agent) report(ctx context.Context) {
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

				err := exporter.SendMetrics(taskCtx, app.config.CollectorAddress, app.config.Secret, app.publicKey, app.stats)

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
	go app.poll(ctx)
	go app.report(ctx)
}
