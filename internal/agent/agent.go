// Package agent implements metrics gathering agent.
package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/alkurbatov/metrics-collector/internal/config"
	"github.com/alkurbatov/metrics-collector/internal/exporter"
	"github.com/alkurbatov/metrics-collector/internal/monitoring"
	"github.com/alkurbatov/metrics-collector/internal/recovery"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/alkurbatov/metrics-collector/internal/validators"
	"github.com/rs/zerolog/log"
)

const _defaultShutdownTimeout = 20 * time.Second

// SendMetrics exports collected metrics in single batch request.
func sendMetrics(
	ctx context.Context,
	exp exporter.Exporter,
	stats *monitoring.Metrics,
) error {
	// NB (alkurbatov): Take snapshot to avoid possible races.
	snapshot := *stats

	exp.
		Add("CPUutilization1", snapshot.System.CPUutilization1).
		Add("TotalMemory", snapshot.System.TotalMemory).
		Add("FreeMemory", snapshot.System.FreeMemory)

	exp.
		Add("Alloc", snapshot.Runtime.Alloc).
		Add("BuckHashSys", snapshot.Runtime.BuckHashSys).
		Add("Frees", snapshot.Runtime.Frees).
		Add("GCCPUFraction", snapshot.Runtime.GCCPUFraction).
		Add("GCSys", snapshot.Runtime.GCSys).
		Add("HeapAlloc", snapshot.Runtime.HeapAlloc).
		Add("HeapIdle", snapshot.Runtime.HeapIdle).
		Add("HeapInuse", snapshot.Runtime.HeapInuse).
		Add("HeapObjects", snapshot.Runtime.HeapObjects).
		Add("HeapReleased", snapshot.Runtime.HeapReleased).
		Add("HeapSys", snapshot.Runtime.HeapSys).
		Add("LastGC", snapshot.Runtime.LastGC).
		Add("Lookups", snapshot.Runtime.Lookups).
		Add("MCacheInuse", snapshot.Runtime.MCacheInuse).
		Add("MCacheSys", snapshot.Runtime.MCacheSys).
		Add("MSpanInuse", snapshot.Runtime.MSpanInuse).
		Add("MSpanSys", snapshot.Runtime.MSpanSys).
		Add("Mallocs", snapshot.Runtime.Mallocs).
		Add("NextGC", snapshot.Runtime.NextGC).
		Add("NumForcedGC", snapshot.Runtime.NumForcedGC).
		Add("NumGC", snapshot.Runtime.NumGC).
		Add("OtherSys", snapshot.Runtime.OtherSys).
		Add("PauseTotalNs", snapshot.Runtime.PauseTotalNs).
		Add("StackInuse", snapshot.Runtime.StackInuse).
		Add("StackSys", snapshot.Runtime.StackSys).
		Add("Sys", snapshot.Runtime.Sys).
		Add("TotalAlloc", snapshot.Runtime.TotalAlloc)

	exp.
		Add("RandomValue", snapshot.RandomValue)

	exp.
		Add("PollCount", snapshot.PollCount)

	if err := exp.Send(ctx).Error(); err != nil {
		return err
	}

	stats.PollCount -= snapshot.PollCount

	return nil
}

type Agent struct {
	// Full configuration of the service.
	config *config.Agent

	// Metrics collected by this Agent.
	stats *monitoring.Metrics

	// Public key used to encrypt agent -> server communications.
	// If the key is nil, communications are not encrypted.
	publicKey security.PublicKey
}

func New(cfg *config.Agent) (*Agent, error) {
	var (
		key security.PublicKey
		err error
	)

	if err = validators.ValidateTransport(cfg.Transport); err != nil {
		return nil, fmt.Errorf("agent - New - validators.ValidateTransport: %w", err)
	}

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

// Poll gathers application and system metrics.
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

// Report sends metrics to the server.
func (app *Agent) report(ctx context.Context) {
	ticker := time.NewTicker(app.config.ReportInterval)
	defer ticker.Stop()

	exp := exporter.New(app.config.Transport, app.config.Address, app.config.Secret, app.publicKey)

	for {
		select {
		case <-ticker.C:
			func() {
				defer recovery.TryRecover()
				defer exp.Reset()

				log.Info().Msg("Sending application metrics")

				// NB (alkurbatov): We have to complete sending data even if shutdown was requested.
				// Thus don't use main context but put timeout.
				taskCtx, cancel := context.WithTimeout(context.Background(), app.config.ExportTimeout)
				defer cancel()

				err := sendMetrics(taskCtx, exp, app.stats)

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

			if err := exp.Close(); err != nil {
				log.Error().Err(err).Msg("agent - report - exp.Close")
			}

			return
		}
	}
}

// Run starts the main app and waits till compeletion or termination signal.
func (app *Agent) Run() {
	ctx, cancelBackgroundTasks := context.WithCancel(context.Background())

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt,
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		app.poll(ctx)
	}()

	go func() {
		defer wg.Done()
		app.report(ctx)
	}()

	s := <-interrupt
	log.Info().Msg("app - Run - interrupt: signal " + s.String())

	log.Info().Msg("Shutting down...")
	cancelBackgroundTasks()

	stopped := make(chan struct{})

	stopCtx, cancel := context.WithTimeout(context.Background(), _defaultShutdownTimeout)
	defer cancel()

	go func() {
		defer close(stopped)
		wg.Wait()
	}()

	select {
	case <-stopped:
		log.Info().Msg("Agent shutdown successful")

	case <-stopCtx.Done():
		log.Warn().Msgf("Exceeded %s shutdown timeout, exit forcibly", _defaultShutdownTimeout)
	}
}
