package app

import (
	"context"
	"time"

	"github.com/alkurbatov/metrics-collector/internal/exporter"
	"github.com/alkurbatov/metrics-collector/internal/logging"
	"github.com/alkurbatov/metrics-collector/internal/metrics"
)

type AgentConfig struct {
	PollInterval     time.Duration
	ReportInterval   time.Duration
	CollectorAddress string
}

func NewAgentConfig() AgentConfig {
	return AgentConfig{
		PollInterval:     2 * time.Second,
		ReportInterval:   10 * time.Second,
		CollectorAddress: "127.0.0.1:8080",
	}
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
			logging.Log.Info("Gathering application metrics")
			stats.Poll()

		case <-ctx.Done():
			logging.Log.Info("Shutdown metrics gatzering")
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
			logging.Log.Info("Sending application metrics")

			if err := exporter.SendMetrics(app.Config.CollectorAddress, *stats); err != nil {
				logging.Log.Error(err)
			}

		case <-ctx.Done():
			logging.Log.Info("Shutdown metrics sending")
			return
		}
	}
}
