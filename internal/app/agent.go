package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/exporter"
	"github.com/alkurbatov/metrics-collector/internal/logging"
	"github.com/alkurbatov/metrics-collector/internal/monitoring"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"
)

type AgentConfig struct {
	PollInterval     time.Duration     `env:"POLL_INTERVAL"`
	ReportInterval   time.Duration     `env:"REPORT_INTERVAL"`
	CollectorAddress entity.NetAddress `env:"ADDRESS"`
	Secret           security.Secret   `env:"KEY"`
	PollTimeout      time.Duration
	ExportTimeout    time.Duration
	Debug            bool `env:"DEBUG"`
}

func NewAgentConfig() AgentConfig {
	collectorAddress := entity.NetAddress("0.0.0.0:8080")
	flag.VarP(
		&collectorAddress,
		"collector-address",
		"a",
		"address:port of metrics collector",
	)

	reportInterval := flag.DurationP(
		"report-interval",
		"r",
		10*time.Second,
		"metrics report interval in seconds",
	)
	pollInterval := flag.DurationP(
		"poll-interval",
		"p",
		2*time.Second,
		"metrics poll interval in seconds",
	)

	secret := security.Secret("")
	flag.VarP(
		&secret,
		"key",
		"k",
		"secret key for signature generation",
	)

	debug := flag.BoolP(
		"debug",
		"g",
		false,
		"enable verbose logging",
	)

	flag.Parse()

	cfg := AgentConfig{
		CollectorAddress: collectorAddress,
		ReportInterval:   *reportInterval,
		PollInterval:     *pollInterval,
		Secret:           secret,
		PollTimeout:      2 * time.Second,
		ExportTimeout:    4 * time.Second,
		Debug:            *debug,
	}

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	return cfg
}

func (c AgentConfig) String() string {
	var sb strings.Builder

	sb.WriteString("Configuration:\n")
	sb.WriteString(fmt.Sprintf("\t\tPoll interval: %s\n", c.PollInterval))
	sb.WriteString(fmt.Sprintf("\t\tReport interval: %s\n", c.ReportInterval))
	sb.WriteString(fmt.Sprintf("\t\tCollector address: %s\n", c.CollectorAddress))

	if len(c.Secret) > 0 {
		sb.WriteString(fmt.Sprintf("\t\tSecret key: %s\n", c.Secret))
	}

	sb.WriteString(fmt.Sprintf("\t\tPollTimeout: %fs\n", c.PollTimeout.Seconds()))
	sb.WriteString(fmt.Sprintf("\t\tExportTimeout: %fs\n", c.ExportTimeout.Seconds()))

	sb.WriteString(fmt.Sprintf("\t\tDebug: %t", c.Debug))

	return sb.String()
}

type Agent struct {
	Config AgentConfig
	stats  monitoring.Metrics
}

func NewAgent() *Agent {
	cfg := NewAgentConfig()

	logging.Setup(cfg.Debug)
	log.Info().Msg(cfg.String())

	return &Agent{Config: cfg}
}

func (app *Agent) poll(ctx context.Context, stats *monitoring.Metrics) {
	ticker := time.NewTicker(app.Config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			func() {
				defer tryRecover()

				log.Info().Msg("Gathering application metrics")

				taskCtx, cancel := context.WithTimeout(ctx, app.Config.PollTimeout)
				defer cancel()

				err := stats.Poll(taskCtx)

				if errors.Is(taskCtx.Err(), context.DeadlineExceeded) {
					log.Error().Dur("deadline", app.Config.PollTimeout).Msg("metrics polling exceeded deadline")
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
	ticker := time.NewTicker(app.Config.ReportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			func() {
				defer tryRecover()

				log.Info().Msg("Sending application metrics")

				taskCtx, cancel := context.WithTimeout(ctx, app.Config.ExportTimeout)
				defer cancel()

				err := exporter.SendMetrics(taskCtx, app.Config.CollectorAddress, app.Config.Secret, stats)

				if errors.Is(err, context.DeadlineExceeded) {
					log.Error().Dur("deadline", app.Config.PollTimeout).Msg("metrics exporting exceeded deadline")
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
