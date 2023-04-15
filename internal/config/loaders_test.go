package config_test

import (
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/config"
	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigFromFile(t *testing.T) {
	tt := []struct {
		name string
		path entity.FilePath
		cfg  config.Config
	}{
		{
			name: "Load agent config",
			path: "../../configs/agent.example.json",
			cfg:  &config.Agent{},
		},
		{
			name: "Load server config",
			path: "../../configs/server.example.json",
			cfg:  &config.Server{},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := config.LoadFromFile(tc.path, tc.cfg)

			require.NoError(t, err)
			snaps.MatchSnapshot(t, tc.cfg.String())
		})
	}
}

func TestLoadUnexistingConfig(t *testing.T) {
	cfg := &config.Agent{}

	err := config.LoadFromFile("xxx", cfg)
	require.Error(t, err)
}
