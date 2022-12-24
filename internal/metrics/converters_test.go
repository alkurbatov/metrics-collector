package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToCounter(t *testing.T) {
	tt := []struct {
		name     string
		value    string
		valid    bool
		expected Counter
	}{
		{
			name:     "Positive integer",
			value:    "15",
			valid:    true,
			expected: 15,
		},
		{
			name:     "Zero integer",
			value:    "0",
			valid:    true,
			expected: 0,
		},
		{
			name:     "Negative integer",
			value:    "-15",
			valid:    true,
			expected: -15,
		},
		{
			name:  "Positive float",
			value: "28968.000000",
			valid: false,
		},
		{
			name:  "Zero float",
			value: "0.000000",
			valid: false,
		},
		{
			name:  "Negative float",
			value: "-28968.000000",
			valid: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			metric, err := ToCounter(tc.value)
			if tc.valid {
				assert.NoError(err)
				assert.Equal(tc.expected, metric)
			} else {
				assert.NotNil(err)
			}
		})
	}
}

func TestToGauge(t *testing.T) {
	tt := []struct {
		name     string
		value    string
		valid    bool
		expected Gauge
	}{
		{
			name:     "Positive integer",
			value:    "15",
			valid:    true,
			expected: 15.0,
		},
		{
			name:     "Zero integer",
			value:    "0",
			valid:    true,
			expected: 0.0,
		},
		{
			name:     "Negative integer",
			value:    "-15",
			valid:    true,
			expected: -15.0,
		},
		{
			name:     "Positive float",
			value:    "28968.134000",
			valid:    true,
			expected: 28968.134,
		},
		{
			name:     "Zero float",
			value:    "0.000000",
			valid:    true,
			expected: 0.0,
		},
		{
			name:     "Negative float",
			value:    "-28968.134000",
			valid:    true,
			expected: -28968.134,
		},
		{
			name:     "Small positive float",
			value:    "0.604660",
			valid:    true,
			expected: 0.604660,
		},
		{
			name:  "Meaningless value",
			value: "...",
			valid: false,
		},
		{
			name:  "Malformed value",
			value: "0.12(",
			valid: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			metric, err := ToGauge(tc.value)
			if tc.valid {
				assert.NoError(err)
				assert.Equal(tc.expected, metric)
			} else {
				assert.NotNil(err)
			}
		})
	}
}
