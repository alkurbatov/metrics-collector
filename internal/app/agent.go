package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alkurbatov/metrics-collector/internal/exporter"
	"github.com/alkurbatov/metrics-collector/internal/logging"
	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/caarlos0/env/v6"
)

type AgentConfig struct {
	PollInterval     time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
	ReportInterval   time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	CollectorAddress string        `env:"ADDRESS" envDefault:"0.0.0.0:8080"`
}

func NewAgentConfig() AgentConfig {
	cfg := AgentConfig{}

	err := env.Parse(&cfg)
	if err != nil {
		logging.Log.Fatal(err)
	}

	return cfg
}

func (c AgentConfig) String() string {
	var sb strings.Builder

	sb.WriteString("Configuration:\n")
	sb.WriteString(fmt.Sprintf("\t\tPoll interval: %s\n", c.PollInterval))
	sb.WriteString(fmt.Sprintf("\t\tReport interval: %s\n", c.ReportInterval))
	sb.WriteString(fmt.Sprintf("\t\tCollector address: %s\n", c.CollectorAddress))

	return sb.String()
}

type Agent struct {
	Config AgentConfig
}

func NewAgent() *Agent {
	return &Agent{Config: NewAgentConfig()}
}

func (app *Agent) Poll(ctx context.Context, stats *metrics.Metrics) {
	ticker := time.NewTicker(app.Config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			func() {
				defer tryRecover()

				logging.Log.Info("Gathering application metrics")
				stats.Poll()
			}()

		case <-ctx.Done():
			logging.Log.Info("Shutdown metrics gathering")
			return
		}
	}
}

func (app *Agent) Report(ctx context.Context, stats *metrics.Metrics) {
	ticker := time.NewTicker(app.Config.ReportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			func() {
				defer tryRecover()

				logging.Log.Info("Sending application metrics")

				if err := exporter.SendMetrics(app.Config.CollectorAddress, *stats); err != nil {
					logging.Log.Error(err)
				}
			}()

		case <-ctx.Done():
			logging.Log.Info("Shutdown metrics sending")
			return
		}
	}
}
