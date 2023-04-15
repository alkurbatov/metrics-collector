package config

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"
)

type Server struct {
	Address        entity.NetAddress    `env:"ADDRESS" json:"address"`
	GRPCAddress    entity.NetAddress    `env:"GRPC_ADDRESS" json:"grpc_address"`
	StoreInterval  time.Duration        `env:"STORE_INTERVAL" json:"store_interval"`
	StorePath      string               `env:"STORE_FILE" json:"store_file"`
	RestoreOnStart bool                 `env:"RESTORE" json:"restore"`
	Secret         security.Secret      `env:"KEY" json:"key"`
	PrivateKeyPath entity.FilePath      `env:"CRYPTO_KEY" json:"crypto_key"`
	TrustedSubnet  *net.IPNet           `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
	DatabaseURL    security.DatabaseURL `env:"DATABASE_DSN" json:"database_dsn"`
	PprofAddress   entity.NetAddress    `env:"PPROF_ADDRESS" json:"pprof_address"`
	Debug          bool                 `env:"DEBUG" json:"debug"`
}

func NewServer() *Server {
	return &Server{
		Address:        "0.0.0.0:8080",
		GRPCAddress:    "0.0.0.0:50051",
		StorePath:      "/tmp/devops-metrics-db.json",
		StoreInterval:  300 * time.Second,
		RestoreOnStart: true,
		Secret:         "",
		PrivateKeyPath: "",
		TrustedSubnet:  nil,
		DatabaseURL:    "",
		Debug:          false,
		PprofAddress:   "",
	}
}

func (c *Server) Parse() error {
	address := c.Address
	flag.VarP(
		&address,
		"address",
		"a",
		"address:port for HTTP API requests",
	)

	grpcAddress := c.Address
	flag.VarP(
		&grpcAddress,
		"grpc-address",
		"s",
		"address:port for gRPC API requests",
	)

	storeInterval := flag.DurationP(
		"store-interval",
		"i",
		c.StoreInterval,
		"count of seconds after which metrics are dumped to the disk, zero value activates saving after each request",
	)
	storePath := flag.StringP(
		"store-file",
		"f",
		c.StorePath,
		"path to file to store metrics",
	)
	restoreOnStart := flag.BoolP(
		"restore",
		"r",
		c.RestoreOnStart,
		"whether to restore state on startup or not",
	)
	secret := c.Secret
	flag.VarP(
		&secret,
		"key",
		"k",
		"secret key for signature generation",
	)

	keyPath := c.PrivateKeyPath
	flag.VarP(
		&keyPath,
		"crypto-key",
		"e",
		"path to private key (stored in PEM format) to decrypt agent -> server communications",
	)

	defaultSubnet := net.IPNet{}
	if c.TrustedSubnet != nil {
		defaultSubnet = *c.TrustedSubnet
	}

	trustedSubnet := flag.IPNetP(
		"trusted-subnet",
		"t",
		defaultSubnet,
		"subnet in CIDR notation, requests from IP address from different subnets will be rejected",
	)

	databaseURL := flag.StringP(
		"db-dsn",
		"d",
		c.DatabaseURL.String(),
		"full database connection URL",
	)

	pprofAddress := c.PprofAddress
	flag.VarP(
		&pprofAddress,
		"pprof-address",
		"p",
		"enable pprof on specified address:port",
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
		if err := loadFromFile(configPath, c); err != nil {
			return err
		}
	}

	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "address":
			c.Address = address

		case "grpc-address":
			c.GRPCAddress = grpcAddress

		case "store-interval":
			c.StoreInterval = *storeInterval

		case "store-file":
			c.StorePath = *storePath

		case "restore":
			c.RestoreOnStart = *restoreOnStart

		case "key":
			c.Secret = secret

		case "crypto-key":
			c.PrivateKeyPath = keyPath

		case "trusted-subnet":
			c.TrustedSubnet = trustedSubnet

		case "db-dsn":
			c.DatabaseURL = security.DatabaseURL(*databaseURL)

		case "pprof-address":
			c.PprofAddress = pprofAddress

		case "debug":
			c.Debug = *debug
		}
	})

	if err := env.Parse(c); err != nil {
		return fmt.Errorf("failed to parse server config: %w", err)
	}

	if len(c.StorePath) == 0 && c.RestoreOnStart {
		return entity.ErrRestoreNoSource
	}

	return nil
}

func (c Server) String() string {
	var sb strings.Builder

	sb.WriteString("Configuration:\n")
	sb.WriteString(fmt.Sprintf("\t\tHTTP API address: %s\n", c.Address))
	sb.WriteString(fmt.Sprintf("\t\tgRPC API address: %s\n", c.GRPCAddress))

	sb.WriteString(fmt.Sprintf("\t\tStore interval: %s\n", c.StoreInterval))
	sb.WriteString(fmt.Sprintf("\t\tStore path: %s\n", c.StorePath))
	sb.WriteString(fmt.Sprintf("\t\tRestore on start: %t\n", c.RestoreOnStart))

	if len(c.Secret) > 0 {
		sb.WriteString(fmt.Sprintf("\t\tSecret key: %s\n", c.Secret))
	}

	if len(c.PrivateKeyPath) > 0 {
		sb.WriteString(fmt.Sprintf("\t\tPrivate key path: %s\n", c.PrivateKeyPath))
	}

	if c.TrustedSubnet != nil {
		sb.WriteString(fmt.Sprintf("\t\tTrusted subnet: %s\n", c.TrustedSubnet.String()))
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
		TrustedSubnet string `json:"trusted_subnet"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("server - UnmarshalJSON - json.Unmarshal: %w", err)
	}

	var err error

	if len(aux.StoreInterval) != 0 {
		c.StoreInterval, err = time.ParseDuration(aux.StoreInterval)
		if err != nil {
			return fmt.Errorf("server - UnmarshalJSON - time.ParseDuration: %w", err)
		}
	}

	if len(aux.TrustedSubnet) != 0 {
		_, c.TrustedSubnet, err = net.ParseCIDR(aux.TrustedSubnet)
		if err != nil {
			return fmt.Errorf("server - UnmarshalJSON - net.ParseCIDR: %w", err)
		}
	}

	return nil
}
