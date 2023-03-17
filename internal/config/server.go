package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"
)

type Server struct {
	ListenAddress  entity.NetAddress    `env:"ADDRESS"`
	StoreInterval  time.Duration        `env:"STORE_INTERVAL"`
	StorePath      string               `env:"STORE_FILE"`
	RestoreOnStart bool                 `env:"RESTORE"`
	Secret         security.Secret      `env:"KEY"`
	DatabaseURL    security.DatabaseURL `env:"DATABASE_DSN"`
	PprofAddress   entity.NetAddress    `env:"PPROF_ADDRESS"`
	Debug          bool                 `env:"DEBUG"`
}

func NewServer() (*Server, error) {
	var (
		listenAddress entity.NetAddress = "0.0.0.0:8080"
		pprofAddress  entity.NetAddress
	)

	flag.VarP(
		&listenAddress,
		"listen-address",
		"a",
		"address:port server listens on",
	)

	storeInterval := flag.DurationP(
		"store-interval",
		"i",
		300*time.Second,
		"count of seconds after which metrics are dumped to the disk, zero value activates saving after each request",
	)
	storePath := flag.StringP(
		"store-file",
		"f",
		"/tmp/devops-metrics-db.json",
		"path to file to store metrics",
	)
	restoreOnStart := flag.BoolP(
		"restore",
		"r",
		true,
		"whether to restore state on startup or not",
	)
	secret := security.Secret("")
	flag.VarP(
		&secret,
		"key",
		"k",
		"secret key for signature generation",
	)

	databaseURL := flag.StringP(
		"db-dsn",
		"d",
		"",
		"full database connection URL",
	)

	flag.VarP(
		&pprofAddress,
		"pprof-address",
		"p",
		"enable pprof on specified address:port",
	)

	debug := flag.BoolP(
		"debug",
		"g",
		false,
		"enable verbose logging",
	)

	flag.Parse()

	cfg := &Server{
		ListenAddress:  listenAddress,
		StorePath:      *storePath,
		StoreInterval:  *storeInterval,
		RestoreOnStart: *restoreOnStart,
		Secret:         secret,
		DatabaseURL:    security.DatabaseURL(*databaseURL),
		Debug:          *debug,
		PprofAddress:   pprofAddress,
	}

	err := env.Parse(cfg)
	if err != nil {
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
	sb.WriteString(fmt.Sprintf("\t\tListening address: %s\n", c.ListenAddress))

	sb.WriteString(fmt.Sprintf("\t\tStore interval: %s\n", c.StoreInterval))
	sb.WriteString(fmt.Sprintf("\t\tStore path: %s\n", c.StorePath))
	sb.WriteString(fmt.Sprintf("\t\tRestore on start: %t\n", c.RestoreOnStart))

	if len(c.Secret) > 0 {
		sb.WriteString(fmt.Sprintf("\t\tSecret key: %s\n", c.Secret))
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
