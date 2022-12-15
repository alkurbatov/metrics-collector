package app

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewAgentConfigWithDefaults(t *testing.T) {
	expected := AgentConfig{
		PollInterval:     2 * time.Second,
		ReportInterval:   10 * time.Second,
		CollectorAddress: "0.0.0.0:8080",
	}

	cfg := NewAgentConfig()

	require.Equal(t, expected, cfg)
}
