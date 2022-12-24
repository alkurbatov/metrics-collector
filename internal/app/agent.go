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
	flag "github.com/spf13/pflag"
)

type AgentConfig struct {
	PollInterval     time.Duration `env:"POLL_INTERVAL"`
	ReportInterval   time.Duration `env:"REPORT_INTERVAL"`
	CollectorAddress NetAddress    `env:"ADDRESS"`
}

func NewAgentConfig() AgentConfig {
	collectorAddress := NetAddress("0.0.0.0:8080")
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
	flag.Parse()

	cfg := AgentConfig{
		CollectorAddress: collectorAddress,
		ReportInterval:   *reportInterval,
		PollInterval:     *pollInterval,
	}

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

				if err := exporter.SendMetrics(app.Config.CollectorAddress.String(), *stats); err != nil {
					logging.Log.Error(err)
				}
			}()

		case <-ctx.Done():
			logging.Log.Info("Shutdown metrics sending")
			return
		}
	}
}
