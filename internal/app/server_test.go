package app

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewServerConfigWithDefaults(t *testing.T) {
	expected := &ServerConfig{
		ListenAddress:  "0.0.0.0:8080",
		StoreInterval:  300 * time.Second,
		StorePath:      "/tmp/devops-metrics-db.json",
		RestoreOnStart: true,
	}

	cfg, err := NewServerConfig()

	require.NoError(t, err)
	require.Equal(t, expected, cfg)
}
