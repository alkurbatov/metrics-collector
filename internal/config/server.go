package config

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"
)

type Server struct {
	Address        entity.NetAddress    `env:"ADDRESS" json:"address"`
	StoreInterval  time.Duration        `env:"STORE_INTERVAL" json:"store_interval"`
	StorePath      string               `env:"STORE_FILE" json:"store_file"`
	RestoreOnStart bool                 `env:"RESTORE" json:"restore"`
	Secret         security.Secret      `env:"KEY" json:"key"`
	PrivateKeyPath entity.FilePath      `env:"CRYPTO_KEY" json:"crypto_key"`
	DatabaseURL    security.DatabaseURL `env:"DATABASE_DSN" json:"database_dsn"`
	PprofAddress   entity.NetAddress    `env:"PPROF_ADDRESS" json:"pprof_address"`
	Debug          bool                 `env:"DEBUG" json:"debug"`
}

func NewServer() (*Server, error) {
	cfg := &Server{
		Address:        "0.0.0.0:8080",
		StorePath:      "/tmp/devops-metrics-db.json",
		StoreInterval:  300 * time.Second,
		RestoreOnStart: true,
		Secret:         "",
		PrivateKeyPath: "",
		DatabaseURL:    "",
		Debug:          false,
		PprofAddress:   "",
	}

	address := cfg.Address
	flag.VarP(
		&address,
		"address",
		"a",
		"address:port server listens on",
	)

	storeInterval := flag.DurationP(
		"store-interval",
		"i",
		cfg.StoreInterval,
		"count of seconds after which metrics are dumped to the disk, zero value activates saving after each request",
	)
	storePath := flag.StringP(
		"store-file",
		"f",
		cfg.StorePath,
		"path to file to store metrics",
	)
	restoreOnStart := flag.BoolP(
		"restore",
		"r",
		cfg.RestoreOnStart,
		"whether to restore state on startup or not",
	)
	secret := cfg.Secret
	flag.VarP(
		&secret,
		"key",
		"k",
		"secret key for signature generation",
	)

	keyPath := cfg.PrivateKeyPath
	flag.VarP(
		&keyPath,
		"crypto-key",
		"e",
		"path to private key (stored in PEM format) to decrypt agent -> server communications",
	)

	databaseURL := flag.StringP(
		"db-dsn",
		"d",
		cfg.DatabaseURL.String(),
		"full database connection URL",
	)

	pprofAddress := cfg.PprofAddress
	flag.VarP(
		&pprofAddress,
		"pprof-address",
		"p",
		"enable pprof on specified address:port",
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

		if f.Name == "store-interval" {
			cfg.StoreInterval = *storeInterval
			return
		}

		if f.Name == "store-file" {
			cfg.StorePath = *storePath
			return
		}

		if f.Name == "restore" {
			cfg.RestoreOnStart = *restoreOnStart
			return
		}

		if f.Name == "key" {
			cfg.Secret = secret
			return
		}

		if f.Name == "crypto-key" {
			cfg.PrivateKeyPath = keyPath
			return
		}

		if f.Name == "db-dsn" {
			cfg.DatabaseURL = security.DatabaseURL(*databaseURL)
			return
		}

		if f.Name == "pprof-address" {
			cfg.PprofAddress = pprofAddress
			return
		}

		if f.Name == "debug" {
			cfg.Debug = *debug
			return
		}
	})

	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse server config: %w", err)
	}

	if len(cfg.StorePath) == 0 && cfg.RestoreOnStart {
		return nil, entity.ErrRestoreNoSource
	}

	return cfg, nil
}

func (c Server) String() string {
	var sb strings.Builder

	sb.WriteString("Configuration:\n")
	sb.WriteString(fmt.Sprintf("\t\tListening address: %s\n", c.Address))

	sb.WriteString(fmt.Sprintf("\t\tStore interval: %s\n", c.StoreInterval))
	sb.WriteString(fmt.Sprintf("\t\tStore path: %s\n", c.StorePath))
	sb.WriteString(fmt.Sprintf("\t\tRestore on start: %t\n", c.RestoreOnStart))

	if len(c.Secret) > 0 {
		sb.WriteString(fmt.Sprintf("\t\tSecret key: %s\n", c.Secret))
	}

	if len(c.PrivateKeyPath) > 0 {
		sb.WriteString(fmt.Sprintf("\t\tPrivate key path: %s\n", c.PrivateKeyPath))
	}

	if len(c.DatabaseURL) > 0 {
		sb.WriteString(fmt.Sprintf("\t\tDatabase URL: %s\n", c.DatabaseURL))
	}

	if len(c.PprofAddress) > 0 {
		sb.WriteString(fmt.Sprintf("\t\tPprof address: %s\n", c.PprofAddress))
	}

	sb.WriteString(fmt.Sprintf("\t\tDebug: %t\n", c.Debug))

	return sb.String()
}

func (c *Server) UnmarshalJSON(data []byte) error {
	type Alias Server

	aux := &struct {
		StoreInterval string `json:"store_interval"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("server - UnmarshalJSON - json.Unmarshal: %w", err)
	}

	storeInterval, err := time.ParseDuration(aux.StoreInterval)
	if err != nil {
		return fmt.Errorf("server - UnmarshalJSON - time.ParseDuration: %w", err)
	}

	c.StoreInterval = storeInterval

	return nil
}
