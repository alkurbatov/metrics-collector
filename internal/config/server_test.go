package config_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/alkurbatov/metrics-collector/internal/config"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestServerConfigString(t *testing.T) {
	tt := []struct {
		name string
		src  *config.Server
	}{
		{
			name: "Test default config to string",
			src:  config.NewServer(),
		},
		{
			name: "Test full config to string",
			src: &config.Server{
				Address:        "0.0.0.0:8080",
				StorePath:      "/tmp/devops-metrics-db.json",
				StoreInterval:  300 * time.Second,
				RestoreOnStart: true,
				Secret:         "xxx",
				PrivateKeyPath: "./keys/key.pem",
				DatabaseURL:    "postgres://postgres:postgres@127.0.0.1:5432/praktikum?sslmode=disable",
				Debug:          false,
				PprofAddress:   "0.0.0.0:3000",
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			snaps.MatchSnapshot(t, tc.src.String())
		})
	}
}

func TestServerConfigUnmarshalJSON(t *testing.T) {
	tt := []struct {
		name     string
		src      string
		expected config.Server
	}{
		{
			name: "Load full configuration",
			src: `{
"address": "0.0.0.0:1234",
"store_interval": "5s",
"store_file": "/tmp/my-db.json",
"restore": true,
"key": "xxx",
"crypto_key": "./keys/key.pem",
"database_dsn": "postgres://postgres:postgres@127.0.0.1:5432/praktikum?sslmode=disable",
"pprof_address": "0.0.0.0:3000",
"debug": true
}`,
			expected: config.Server{
				Address:        "0.0.0.0:1234",
				StorePath:      "/tmp/my-db.json",
				StoreInterval:  5 * time.Second,
				RestoreOnStart: true,
				Secret:         "xxx",
				PrivateKeyPath: "./keys/key.pem",
				DatabaseURL:    "postgres://postgres:postgres@127.0.0.1:5432/praktikum?sslmode=disable",
				Debug:          true,
				PprofAddress:   "0.0.0.0:3000",
			},
		},
		{
			name: "Load partial configuration",
			src: `{
"crypto_key": "./keys/key.pem",
"debug": false
}`,
			expected: config.Server{
				Address:        "0.0.0.0:8080",
				StorePath:      "/tmp/devops-metrics-db.json",
				StoreInterval:  300 * time.Second,
				RestoreOnStart: true,
				Secret:         "",
				PrivateKeyPath: "./keys/key.pem",
				DatabaseURL:    "",
				Debug:          false,
				PprofAddress:   "",
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.NewServer()

			err := json.Unmarshal([]byte(tc.src), cfg)
			require.NoError(t, err)

			require.Equal(t, tc.expected, *cfg)
		})
	}
}

func TestServerConfigUnmarshallInvalidJSON(t *testing.T) {
	tt := []struct {
		name string
		src  string
	}{
		{
			name: "Parse config with invalid data",
			src: `{
"address": 2
}`,
		},
		{
			name: "Parse config with invalid store interval",
			src: `{
"store_interval": "_"
}`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.NewServer()

			err := json.Unmarshal([]byte(tc.src), cfg)
			require.Error(t, err)
		})
	}
}
