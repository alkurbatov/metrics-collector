package config

import (
	"encoding/json"
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
	PollInterval   time.Duration     `env:"POLL_INTERVAL" json:"poll_interval"`
	ReportInterval time.Duration     `env:"REPORT_INTERVAL" json:"report_interval"`
	Address        entity.NetAddress `env:"ADDRESS" json:"address"`
	Secret         security.Secret   `env:"KEY" json:"key"`
	PublicKeyPath  entity.FilePath   `env:"CRYPTO_KEY" json:"crypto_key"`
	PollTimeout    time.Duration
	ExportTimeout  time.Duration
	Debug          bool `env:"DEBUG" json:"debug"`
}

func NewAgent() (*Agent, error) {
	cfg := Agent{
		Address:        "0.0.0.0:8080",
		ReportInterval: 10 * time.Second,
		PollInterval:   2 * time.Second,
		Secret:         "",
		PublicKeyPath:  "",
		PollTimeout:    2 * time.Second,
		ExportTimeout:  4 * time.Second,
		Debug:          false,
	}

	address := cfg.Address
	flag.VarP(
		&address,
		"collector-address",
		"a",
		"address:port of metrics collector",
	)

	reportInterval := flag.DurationP(
		"report-interval",
		"r",
		cfg.ReportInterval,
		"metrics report interval in seconds",
	)
	pollInterval := flag.DurationP(
		"poll-interval",
		"p",
		cfg.PollInterval,
		"metrics poll interval in seconds",
	)

	secret := cfg.Secret
	flag.VarP(
		&secret,
		"key",
		"k",
		"secret key for signature generation",
	)

	keyPath := cfg.PublicKeyPath
	flag.VarP(
		&keyPath,
		"crypto-key",
		"e",
		"path to public key (stored in PEM format) to encrypt agent -> server communications",
	)

	debug := flag.BoolP(
		"debug",
		"g",
		cfg.Debug,
		"enable verbose logging",
	)

	configPath := entity.FilePath("")
	flag.VarP(
		&configPath,
		"config",
		"c",
		"path to configuration file in JSON format",
	)

	flag.Parse()

	if len(configPath) != 0 {
		if err := loadFromFile(configPath, &cfg); err != nil {
			return nil, err
		}
	}

	flag.Visit(func(f *flag.Flag) {
		if f.Name == "address" {
			cfg.Address = address
			return
		}

		if f.Name == "report-interval" {
			cfg.ReportInterval = *reportInterval
			return
		}

		if f.Name == "poll-interval" {
			cfg.PollInterval = *pollInterval
			return
		}

		if f.Name == "key" {
			cfg.Secret = secret
			return
		}

		if f.Name == "crypto-key" {
			cfg.PublicKeyPath = keyPath
			return
		}

		if f.Name == "debug" {
			cfg.Debug = *debug
			return
		}
	})

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	return &cfg, nil
}

func (c Agent) String() string {
	var sb strings.Builder

	sb.WriteString("Configuration:\n")
	sb.WriteString(fmt.Sprintf("\t\tPoll interval: %s\n", c.PollInterval))
	sb.WriteString(fmt.Sprintf("\t\tReport interval: %s\n", c.ReportInterval))
	sb.WriteString(fmt.Sprintf("\t\tCollector address: %s\n", c.Address))

	if len(c.Secret) > 0 {
		sb.WriteString(fmt.Sprintf("\t\tSecret key: %s\n", c.Secret))
	}

	if len(c.PublicKeyPath) > 0 {
		sb.WriteString(fmt.Sprintf("\t\tPublic key path: %s\n", c.PublicKeyPath))
	}

	sb.WriteString(fmt.Sprintf("\t\tPollTimeout: %fs\n", c.PollTimeout.Seconds()))
	sb.WriteString(fmt.Sprintf("\t\tExportTimeout: %fs\n", c.ExportTimeout.Seconds()))

	sb.WriteString(fmt.Sprintf("\t\tDebug: %t", c.Debug))

	return sb.String()
}

func (c *Agent) UnmarshalJSON(data []byte) error {
	type Alias Agent

	aux := &struct {
		PollInterval   string `json:"poll_interval"`
		ReportInterval string `json:"report_interval"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("agent - UnmarshalJSON - json.Unmarshal: %w", err)
	}

	reportInterval, err := time.ParseDuration(aux.ReportInterval)
	if err != nil {
		return fmt.Errorf("agent - UnmarshalJSON - time.ParseDuration: %w", err)
	}

	pollInterval, err := time.ParseDuration(aux.PollInterval)
	if err != nil {
		return fmt.Errorf("agent - UnmarshalJSON - time.ParseDuration: %w", err)
	}

	c.ReportInterval = reportInterval
	c.PollInterval = pollInterval

	return nil
}
