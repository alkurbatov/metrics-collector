package security_test

import (
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/stretchr/testify/assert"
)

func TestSecretStringConversion(t *testing.T) {
	tt := []struct {
		name     string
		secret   security.Secret
		expected string
	}{
		{
			name:     "Should hide content",
			secret:   "some-secret",
			expected: "***********",
		},
		{
			name:     "Should work with empty content",
			secret:   "",
			expected: "",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.secret.String())
		})
	}
}

func TestDatabaseURLStringConversion(t *testing.T) {
	tt := []struct {
		name     string
		url      security.DatabaseURL
		expected string
	}{
		{
			name:     "Should hide login and password",
			url:      "postgres://postgres:postgres@127.0.0.1:5432/praktikum?sslmode=disable",
			expected: "postgres://*****:*****@127.0.0.1:5432/praktikum?sslmode=disable",
		},
		{
			name:     "Should work with empty content",
			url:      "",
			expected: "",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.url.String())
		})
	}
}
