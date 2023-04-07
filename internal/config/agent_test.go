package config_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/alkurbatov/metrics-collector/internal/config"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestAgentConfigString(t *testing.T) {
	tt := []struct {
		name string
		src  *config.Agent
	}{
		{
			name: "Test default config to string",
			src:  config.NewAgent(),
		},
		{
			name: "Test full config to string",
			src: &config.Agent{
				Address:        "0.0.0.0:8080",
				PollInterval:   2 * time.Second,
				ReportInterval: 10 * time.Second,
				Secret:         "xxx",
				PublicKeyPath:  "./keys/key.pem",
				PollTimeout:    2 * time.Second,
				ExportTimeout:  4 * time.Second,
				Debug:          false,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			snaps.MatchSnapshot(t, tc.src.String())
		})
	}
}

func TestAgentConfigUnmarshalJSON(t *testing.T) {
	tt := []struct {
		name     string
		src      string
		expected config.Agent
	}{
		{
			name: "Load full configuration",
			src: `{
"address": "0.0.0.0:1234",
"poll_interval": "2s",
"report_interval": "5s",
"key": "xxx",
"crypto_key": "./keys/key.pem",
"debug": true
}`,
			expected: config.Agent{
				Address:        "0.0.0.0:1234",
				PollInterval:   2 * time.Second,
				ReportInterval: 5 * time.Second,
				Secret:         "xxx",
				PublicKeyPath:  "./keys/key.pem",
				PollTimeout:    2 * time.Second,
				ExportTimeout:  4 * time.Second,
				Debug:          true,
			},
		},
		{
			name: "Load partial configuration",
			src: `{
"crypto_key": "./keys/key.pem",
"debug": false
}`,
			expected: config.Agent{
				Address:        "0.0.0.0:8080",
				PollInterval:   2 * time.Second,
				ReportInterval: 10 * time.Second,
				Secret:         "",
				PublicKeyPath:  "./keys/key.pem",
				PollTimeout:    2 * time.Second,
				ExportTimeout:  4 * time.Second,
				Debug:          false,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.NewAgent()

			err := json.Unmarshal([]byte(tc.src), cfg)
			require.NoError(t, err)

			require.Equal(t, tc.expected, *cfg)
		})
	}
}

func TestAgentConfigUnmarshallInvalidJSON(t *testing.T) {
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
			name: "Parse config with invalid poll interval",
			src: `{
"poll_interval": "_"
}`,
		},
		{
			name: "Parse config with invalid report interval",
			src: `{
"report_interval": "x"
}`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.NewAgent()

			err := json.Unmarshal([]byte(tc.src), cfg)
			require.Error(t, err)
		})
	}
}
