package app

import (
	"github.com/alkurbatov/metrics-collector/internal/storage"
)

type ServerConfig struct {
	ListenAddress string
}

func NewServerConfig() ServerConfig {
	return ServerConfig{
		ListenAddress: "0.0.0.0:8080",
	}
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
