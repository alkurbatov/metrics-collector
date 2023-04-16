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
	Transport      string            `env:"TRANSPORT" json:"transport"`
	Secret         security.Secret   `env:"KEY" json:"key"`
	PublicKeyPath  entity.FilePath   `env:"CRYPTO_KEY" json:"crypto_key"`
	PollTimeout    time.Duration     `json:"-"`
	ExportTimeout  time.Duration     `json:"-"`
	Debug          bool              `env:"DEBUG" json:"debug"`
}

func NewAgent() *Agent {
	return &Agent{
		Address:        "0.0.0.0:8080",
		Transport:      entity.TransportHTTP,
		ReportInterval: 10 * time.Second,
		PollInterval:   2 * time.Second,
		Secret:         "",
		PublicKeyPath:  "",
		PollTimeout:    2 * time.Second,
		ExportTimeout:  4 * time.Second,
		Debug:          false,
	}
}

func (c *Agent) Parse() error {
	address := c.Address
	flag.VarP(
		&address,
		"address",
		"a",
		"address:port of metrics collector",
	)

	transport := flag.StringP(
		"transport",
		"t",
		c.Transport,
		"type of transport used to export metrics to the server (http or grpc)",
	)

	reportInterval := flag.DurationP(
		"report-interval",
		"r",
		c.ReportInterval,
		"metrics report interval in seconds",
	)
	pollInterval := flag.DurationP(
		"poll-interval",
		"p",
		c.PollInterval,
		"metrics poll interval in seconds",
	)

	secret := c.Secret
	flag.VarP(
		&secret,
		"key",
		"k",
		"secret key for signature generation",
	)

	keyPath := c.PublicKeyPath
	flag.VarP(
		&keyPath,
		"crypto-key",
		"e",
		"path to public key (stored in PEM format) to encrypt agent -> server communications",
	)

	debug := flag.BoolP(
		"debug",
		"g",
		c.Debug,
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
		if err := LoadFromFile(configPath, c); err != nil {
			return err
		}
	}

	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "address":
			c.Address = address

		case "transport":
			c.Transport = *transport

		case "report-interval":
			c.ReportInterval = *reportInterval

		case "poll-interval":
			c.PollInterval = *pollInterval

		case "key":
			c.Secret = secret

		case "crypto-key":
			c.PublicKeyPath = keyPath

		case "debug":
			c.Debug = *debug
		}
	})

	err := env.Parse(c)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	c.Transport = strings.ToLower(c.Transport)

	return nil
}

func (c Agent) String() string {
	var sb strings.Builder

	sb.WriteString("Configuration:\n")
	sb.WriteString(fmt.Sprintf("\t\tPoll interval: %s\n", c.PollInterval))
	sb.WriteString(fmt.Sprintf("\t\tReport interval: %s\n", c.ReportInterval))
	sb.WriteString(fmt.Sprintf("\t\tCollector address: %s\n", c.Address))
	sb.WriteString(fmt.Sprintf("\t\tTransport: %s\n", c.Transport))

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

	var err error

	if len(aux.ReportInterval) != 0 {
		c.ReportInterval, err = time.ParseDuration(aux.ReportInterval)
		if err != nil {
			return fmt.Errorf("agent - UnmarshalJSON - time.ParseDuration: %w", err)
		}
	}

	if len(aux.PollInterval) != 0 {
		c.PollInterval, err = time.ParseDuration(aux.PollInterval)
		if err != nil {
			return fmt.Errorf("agent - UnmarshalJSON - time.ParseDuration: %w", err)
		}
	}

	return nil
}
