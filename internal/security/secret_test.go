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
			name:     "Basic secret",
			secret:   "some-secret",
			expected: "***********",
		},
		{
			name:     "Empty secret",
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
