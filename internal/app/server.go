package app

import (
	"fmt"
	"strings"

	"github.com/alkurbatov/metrics-collector/internal/logging"
	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	ListenAddress string `env:"ADDRESS" envDefault:"0.0.0.0:8080"`
}

func NewServerConfig() ServerConfig {
	cfg := ServerConfig{}

	err := env.Parse(&cfg)
	if err != nil {
		logging.Log.Fatal(err)
	}

	return cfg
}

func (c ServerConfig) String() string {
	var sb strings.Builder

	sb.WriteString("Configuration:\n")
	sb.WriteString(fmt.Sprintf("\t\tListening address: %s\n", c.ListenAddress))

	return sb.String()
}

type Server struct {
	Config  ServerConfig
	Storage storage.Storage
}

func NewServer() *Server {
	return &Server{
		Config:  NewServerConfig(),
		Storage: storage.NewMemStorage(),
	}
}
