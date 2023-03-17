package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"
)

type Agent struct {
	PollInterval     time.Duration     `env:"POLL_INTERVAL"`
	ReportInterval   time.Duration     `env:"REPORT_INTERVAL"`
	CollectorAddress entity.NetAddress `env:"ADDRESS"`
	Secret           security.Secret   `env:"KEY"`
	PollTimeout      time.Duration
	ExportTimeout    time.Duration
	Debug            bool `env:"DEBUG"`
}

func NewAgent() *Agent {
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

	cfg := Agent{
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

	return &cfg
}

func (c Agent) String() string {
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
